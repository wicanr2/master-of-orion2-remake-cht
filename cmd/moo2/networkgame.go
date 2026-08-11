package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"net"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/netplay"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

var (
	colorDarkNetwork  = color.RGBA{6, 8, 14, 255}
	colorPanelNetwork = color.RGBA{12, 18, 32, 245}
	colorGoldNetwork  = color.RGBA{240, 220, 120, 255}
	colorBodyNetwork  = color.RGBA{214, 222, 238, 255}
	colorDimNetwork   = color.RGBA{150, 162, 185, 255}
	colorErrorNetwork = color.RGBA{235, 140, 130, 255}
)

// networkgame.go：把 TCP 大廳接到真正的共同開局與決定性鎖步回合。
//
// 網路上只傳兩種東西：主機選定的完整快照，以及每位玩家本回合的
// PlayerCommand 序列。所有機器都從同一份回合基準快照重播，重播後再交換
// 第二階段指紋；指紋不一致就失敗即關閉，不能讓分岔的對局繼續假裝正常。

type networkStartPayload struct {
	Roster   netplay.Roster `json:"roster"`
	Snapshot []byte         `json:"snapshot"`
}

type networkTurnState struct {
	base      []byte
	table     *netplay.Table
	commands  []shell.PlayerCommand
	submitted bool
	replayed  bool
	readySent bool
	errorText string
	chat      []string
}

// beginNetworkHostSetup 是主機在等待面板按下 START NET GAME 後的入口。
func (b *sceneBuilder) beginNetworkHostSetup() *origTransition {
	if b.netLobby == nil {
		return nil
	}
	r := b.netLobby.Roster()
	if len(r.Players) < 2 {
		return nil // 尚未有第二位玩家，按鈕保持等待狀態。
	}
	b.networkRoster = r
	b.networkPending = true
	b.pendingHotseat = 0
	b.pendingHotseatAI = nil
	if len(r.Players) > b.newGameEmpires {
		b.newGameEmpires = len(r.Players)
	}
	return b.goTo(b.newGameSetup, "網路共同開局設定")
}

// finishNetworkHostSetup 在主機完成種族、旗色與名稱後廣播唯一的共同開局。
func (b *sceneBuilder) finishNetworkHostSetup() *origTransition {
	if b.session == nil || b.netLobby == nil {
		return nil
	}
	r := b.netLobby.Roster()
	if len(r.Players) < 2 {
		return nil
	}
	names := make([]string, len(r.Players))
	for _, p := range r.Players {
		if p.ID >= 0 && p.ID < len(names) {
			names[p.ID] = p.Name
		}
	}
	for i := range names {
		if names[i] == "" {
			names[i] = fmt.Sprintf("Player %d", i+1)
		}
	}
	if got := b.session.SetupNetworkWithNames(names); got != len(names) {
		return nil
	}
	if err := b.session.SetActiveSeat(0); err != nil {
		return nil
	}
	snapshot, err := b.session.MarshalSnapshot()
	if err != nil {
		return nil
	}
	payload, err := json.Marshal(networkStartPayload{Roster: r, Snapshot: snapshot})
	if err != nil {
		return nil
	}
	if err := b.netLobby.Broadcast(netplay.Message{
		Kind:    netplay.KindGameStart,
		Turn:    b.session.Turn,
		Payload: string(payload),
	}); err != nil {
		return nil
	}
	if b.netAnnouncer != nil {
		_ = b.netAnnouncer.Close()
		b.netAnnouncer = nil
	}
	b.networkRoster = r
	// 保留 listen socket，只拒絕新身份；舊玩家可用 lobby resume token
	// 回來，Lobby callback 再把新連線交給同一個 Session peer id。
	sess := netplay.NewSessionWithOptions(0, true, b.netLobby.Connections(), b.netSessionOptions())
	b.netSess = sess
	b.netLobby.SetReconnectHandler(func(p netplay.Player, c net.Conn) {
		_ = sess.ReplacePeer(p.ID, c)
	})
	b.netLobby.SetReconnectOnly()
	b.netConn = nil
	b.networkPending = false
	if err := b.startNetworkTurnState(); err != nil {
		b.networkError = err.Error()
		return nil
	}
	return b.goTo(b.galaxy, "網路星系主畫面")
}

// acceptNetworkGame 套用主機的快照；客戶端不得自行重新擲骰星系。
func (b *sceneBuilder) acceptNetworkGame(raw string) (*origTransition, error) {
	var start networkStartPayload
	if err := json.Unmarshal([]byte(raw), &start); err != nil {
		return nil, fmt.Errorf("解析主機共同開局：%w", err)
	}
	if len(start.Roster.Players) < 2 || len(start.Snapshot) == 0 {
		return nil, fmt.Errorf("主機共同開局資料不完整")
	}
	gs, err := shell.RestoreSnapshot(start.Snapshot)
	if err != nil {
		return nil, err
	}
	if err := gs.SetActiveSeat(b.netMe); err != nil {
		return nil, err
	}
	b.session = gs
	b.networkRoster = start.Roster
	b.networkPending = false
	b.applyNebulaStarFlags(b.session)
	if len(b.herodataMercs) > 0 {
		b.session.SetMercCandidates(b.herodataMercs)
	}
	if err := b.startNetworkTurnState(); err != nil {
		return nil, err
	}
	return b.goTo(b.galaxy, "網路星系主畫面"), nil
}

// startNetworkTurnState 建立本回合的共同基準，並開啟玩家操作記錄。
func (b *sceneBuilder) startNetworkTurnState() error {
	if b.session == nil || b.netSess == nil {
		return fmt.Errorf("網路對局缺少 session")
	}
	if len(b.networkRoster.Players) < 2 {
		return fmt.Errorf("網路對局玩家不足")
	}
	b.networkTurn = &networkTurnState{}
	b.session.SetCommandRecorder(nil)
	if err := b.session.SetActiveSeat(0); err != nil {
		return err
	}
	base, err := b.session.MarshalSnapshot()
	if err != nil {
		return err
	}
	b.networkTurn.base = base
	if err := b.session.SetActiveSeat(b.netMe); err != nil {
		return err
	}
	b.installNetworkRecorder()
	return nil
}

func (b *sceneBuilder) installNetworkRecorder() {
	if b.session == nil {
		return
	}
	b.session.SetCommandRecorder(func(c shell.PlayerCommand) {
		st := b.networkTurn
		if st == nil || st.submitted || st.replayed || st.errorText != "" {
			return
		}
		st.commands = append(st.commands, c)
	})
}

func toNetCommand(c shell.PlayerCommand) netplay.Command {
	return netplay.Command{Name: c.Name, Args: append([]int(nil), c.Args...), Text: c.Text}
}

func toShellCommand(c netplay.Command) shell.PlayerCommand {
	return shell.PlayerCommand{Name: c.Name, Args: append([]int(nil), c.Args...), Text: c.Text}
}

// submitNetworkTurn 將本地玩家本回合的指令送出；世界回合尚未推進。
func (b *sceneBuilder) submitNetworkTurn() *origTransition {
	st := b.networkTurn
	if st == nil || b.netSess == nil || b.session == nil {
		return nil
	}
	if st.submitted {
		return &origTransition{next: b.networkWaitScreen()}
	}
	st.table = netplay.NewTable(len(b.networkRoster.Players), b.session.Turn)
	commands := make([]netplay.Command, 0, len(st.commands))
	for _, c := range st.commands {
		commands = append(commands, toNetCommand(c))
	}
	m := netplay.Message{Kind: netplay.KindTurnDone, Turn: b.session.Turn, Commands: commands}
	if err := st.table.Add(mWithPlayer(m, b.netMe)); err != nil {
		b.failNetwork(err.Error())
		return &origTransition{next: b.networkWaitScreen()}
	}
	if err := b.netSess.Send(m); err != nil {
		b.failNetwork(b.tr("送出回合指令：", "Send turn commands: ") + err.Error())
	}
	st.submitted = true
	return &origTransition{next: b.networkWaitScreen()}
}

func mWithPlayer(m netplay.Message, player int) netplay.Message {
	m.Player = player
	return m
}

type networkWaitScreen struct {
	b    *sceneBuilder
	tick int
}

func (b *sceneBuilder) networkWaitScreen() *networkWaitScreen {
	return &networkWaitScreen{b: b}
}

func (s *networkWaitScreen) update(in shell.InputState) *origTransition {
	s.tick++
	st := s.b.networkTurn
	if st == nil {
		return nil
	}
	if st.errorText == "" && s.b.netSess != nil {
		for _, m := range s.b.netSess.Poll() {
			s.handle(m)
		}
		if err := s.b.netSess.Err(); err != nil {
			s.b.failNetwork(s.b.tr("網路連線錯誤：", "Network error: ") + err.Error())
		}
	}
	if st.errorText == "" && st.table != nil && st.table.Ready() && !st.replayed {
		if err := s.replayCommands(); err != nil {
			s.b.failNetwork(err.Error())
		}
	}
	if st.errorText == "" && st.readySent && st.table.ReadyComplete() {
		if detail := st.table.ReadyDesync(); detail != "" {
			s.sendDesync(detail)
			s.b.failNetwork(detail)
		} else {
			return s.resolveTurn()
		}
	}
	if st.errorText != "" && in.ClickReleased {
		s.b.closeNetwork()
		sc, err := s.b.multiPlayer()
		if err == nil {
			return &origTransition{next: sc}
		}
	}
	return nil
}

func (s *networkWaitScreen) handle(m netplay.Message) {
	st := s.b.networkTurn
	if st == nil {
		return
	}
	switch m.Kind {
	case netplay.KindTurnDone:
		if st.table == nil {
			s.b.failNetwork(s.b.tr("尚未建立回合表就收到玩家指令", "Received commands before the turn table was ready"))
			return
		}
		if err := st.table.Add(m); err != nil {
			s.b.failNetwork(err.Error())
		}
	case netplay.KindTurnReady:
		if st.table == nil {
			s.b.failNetwork(s.b.tr("尚未建立回合表就收到 ready", "Received ready before the turn table was ready"))
			return
		}
		if err := st.table.AddReady(m); err != nil {
			s.b.failNetwork(err.Error())
		}
	case netplay.KindDesync:
		s.b.failNetwork(m.Detail)
	case netplay.KindChat:
		if m.Text != "" {
			st.chat = append(st.chat, fmt.Sprintf("P%d: %s", m.Player+1, m.Text))
			if len(st.chat) > 4 {
				st.chat = st.chat[len(st.chat)-4:]
			}
		}
	}
}

func (s *networkWaitScreen) replayCommands() error {
	st := s.b.networkTurn
	if st == nil || st.table == nil {
		return fmt.Errorf("網路回合資料不存在")
	}
	gs, err := shell.RestoreSnapshot(st.base)
	if err != nil {
		return err
	}
	s.b.session = gs
	s.b.session.SetCommandRecorder(nil)
	if err := s.b.session.SetActiveSeat(0); err != nil {
		return err
	}
	for p := 0; p < len(s.b.networkRoster.Players); p++ {
		wire := st.table.CommandsFor(p)
		cmds := make([]shell.PlayerCommand, 0, len(wire))
		for _, c := range wire {
			cmds = append(cmds, toShellCommand(c))
		}
		if err := s.b.session.ApplyPlayerCommandsForSeat(p, cmds); err != nil {
			return fmt.Errorf("重播玩家 %d 指令：%w", p, err)
		}
	}
	if err := s.b.session.SetActiveSeat(0); err != nil {
		return err
	}
	hash := s.b.session.NetworkStateHash()
	if hash == "" {
		return fmt.Errorf("重播後算不出狀態指紋")
	}
	ready := netplay.Message{Kind: netplay.KindTurnReady, Turn: st.table.Turn(), StateHash: hash}
	if err := st.table.AddReady(mWithPlayer(ready, s.b.netMe)); err != nil {
		return err
	}
	if err := s.b.netSess.Send(ready); err != nil {
		return fmt.Errorf("送出重播指紋：%w", err)
	}
	st.replayed = true
	st.readySent = true
	return nil
}

func (s *networkWaitScreen) sendDesync(detail string) {
	if s.b.netSess == nil {
		return
	}
	_ = s.b.netSess.Send(netplay.Message{Kind: netplay.KindDesync, Player: s.b.netMe, Detail: detail})
}

func (s *networkWaitScreen) resolveTurn() *origTransition {
	st := s.b.networkTurn
	if st == nil || s.b.session == nil {
		return nil
	}
	s.b.session.SetCommandRecorder(nil)
	if err := s.b.session.SetActiveSeat(0); err != nil {
		s.b.failNetwork(err.Error())
		return nil
	}
	s.b.session.EndTurn()
	if err := s.b.captureNetworkTurnBase(); err != nil {
		s.b.failNetwork(err.Error())
		return nil
	}
	return s.b.finishResolvedTurn()
}

func (b *sceneBuilder) captureNetworkTurnBase() error {
	if b.networkTurn == nil || b.session == nil {
		return fmt.Errorf("網路回合狀態不存在")
	}
	b.session.SetCommandRecorder(nil)
	if err := b.session.SetActiveSeat(0); err != nil {
		return err
	}
	base, err := b.session.MarshalSnapshot()
	if err != nil {
		return err
	}
	b.networkTurn.base = base
	b.networkTurn.table = nil
	b.networkTurn.commands = nil
	b.networkTurn.submitted = false
	b.networkTurn.replayed = false
	b.networkTurn.readySent = false
	if err := b.session.SetActiveSeat(b.netMe); err != nil {
		return err
	}
	b.installNetworkRecorder()
	return nil
}

func (b *sceneBuilder) failNetwork(detail string) {
	if detail == "" {
		detail = b.tr("網路對局已中止", "Network match stopped")
	}
	if b.networkTurn != nil {
		b.networkTurn.errorText = detail
	}
	b.networkError = detail
	if b.session != nil {
		b.session.SetCommandRecorder(nil)
	}
	if b.netSess != nil {
		_ = b.netSess.Close()
	}
}

func (b *sceneBuilder) closeNetwork() {
	if b.session != nil {
		b.session.SetCommandRecorder(nil)
	}
	if b.netSess != nil {
		_ = b.netSess.Close()
	}
	if b.netLobby != nil {
		_ = b.netLobby.Close()
	}
	if b.netAnnouncer != nil {
		_ = b.netAnnouncer.Close()
	}
	b.netSess = nil
	b.netLobby = nil
	b.netConn = nil
	b.netAddr = ""
	b.netPlayerName = ""
	b.netJoinOptions = netplay.JoinOptions{}
	b.netLobbyOpts = netplay.LobbyOptions{}
	b.netAnnouncer = nil
	b.networkTurn = nil
	b.networkHost = false
	b.networkPending = false
	b.networkError = ""
}

func (s *networkWaitScreen) draw(dst *ebiten.Image) {
	dst.Fill(colorDarkNetwork)
	if s.b.fnt == nil {
		return
	}
	st := s.b.networkTurn
	fillPanel(dst, 42, 42, 556, 396, colorPanelNetwork, false)
	title := s.b.tr("網路回合鎖步", "NETWORK TURN LOCKSTEP")
	s.b.fnt.DrawCentered(dst, title, 320, 70, 18, colorGoldNetwork)
	if st == nil {
		return
	}
	if st.errorText != "" {
		s.b.fnt.DrawCentered(dst, s.b.tr("對局已停止：", "GAME STOPPED: ")+st.errorText,
			320, 120, 13, colorErrorNetwork)
		s.b.fnt.DrawCentered(dst, s.b.tr("點擊返回多人設定", "Click to return to multiplayer setup"),
			320, 410, 12, colorBodyNetwork)
		return
	}
	turn := 0
	if st.table != nil {
		turn = st.table.Turn()
	} else if s.b.session != nil {
		turn = s.b.session.Turn
	}
	s.b.fnt.DrawCentered(dst, fmt.Sprintf(s.b.tr("第 %d 回合", "TURN %d"), turn), 320, 102, 14, colorBodyNetwork)
	status := s.b.tr("等待所有玩家完成指令…", "Waiting for all players to submit commands…")
	if st.table != nil && st.table.Ready() {
		status = s.b.tr("指令已到齊，正在比對重播結果…", "Commands received; comparing replay results…")
	} else if st.submitted {
		status = s.b.tr("你已提交，等待其他玩家…", "Submitted; waiting for other players…")
	}
	s.b.fnt.DrawCentered(dst, status, 320, 130, 12, colorGoldNetwork)
	for i, p := range s.b.networkRoster.Players {
		name := truncateToWidth(s.b.fnt, p.Name, 12, 190)
		state := s.b.tr("未提交", "not ready")
		if st.table != nil {
			if st.table.Ready() {
				state = s.b.tr("已重播", "replayed")
			} else if !containsInt(st.table.Missing(), p.ID) {
				state = s.b.tr("已提交", "submitted")
			}
		}
		s.b.fnt.Draw(dst, fmt.Sprintf("%d. %s", i+1, name), 78, float64(165+i*28), 12, colorBodyNetwork)
		s.b.fnt.Draw(dst, state, 390, float64(165+i*28), 12, colorDimNetwork)
	}
	for i, line := range st.chat {
		s.b.fnt.Draw(dst, truncateToWidth(s.b.fnt, line, 10, 490), 74, float64(335+i*15), 10, colorDimNetwork)
	}
}

func containsInt(values []int, want int) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}
