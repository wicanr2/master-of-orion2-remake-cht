package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"net"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/netplay"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

var (
	colorDarkNetwork  = color.RGBA{6, 8, 14, 255}
	colorBodyNetwork  = color.RGBA{214, 222, 238, 255}
	colorErrorNetwork = color.RGBA{235, 140, 130, 255}
)

// networkgame.go：把 TCP 大廳接到真正的共同開局與決定性鎖步回合。
//
// 決定性遊戲狀態只由兩種資料形成：主機選定的完整快照，以及每位玩家本回合的
// PlayerCommand 序列。聊天、心跳、重連與分岔通知是控制訊息，不直接修改遊戲狀態。
// 所有機器都從同一份回合基準快照重播，重播後再交換第二階段指紋；指紋不一致就
// 失敗即關閉，不能讓分岔的對局繼續假裝正常。

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
	return b.goTo(b.newGameSetup, uiText(b.lang, "network.transition.shared_setup"))
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
			names[i] = fmt.Sprintf(uiText(b.lang, "network.player.fallback"), i+1)
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
	return b.goTo(b.galaxy, uiText(b.lang, "network.transition.galaxy"))
}

// acceptNetworkGame 套用主機的快照；客戶端不得自行重新擲骰星系。
func (b *sceneBuilder) acceptNetworkGame(raw string) (*origTransition, error) {
	var start networkStartPayload
	if err := json.Unmarshal([]byte(raw), &start); err != nil {
		return nil, fmt.Errorf(uiText(b.lang, "network.error.parse_start"), err)
	}
	if len(start.Roster.Players) < 2 || len(start.Snapshot) == 0 {
		return nil, fmt.Errorf("%s", uiText(b.lang, "network.error.incomplete_start"))
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
	return b.goTo(b.galaxy, uiText(b.lang, "network.transition.galaxy")), nil
}

// startNetworkTurnState 建立本回合的共同基準，並開啟玩家操作記錄。
func (b *sceneBuilder) startNetworkTurnState() error {
	if b.session == nil || b.netSess == nil {
		return fmt.Errorf("%s", uiText(b.lang, "network.error.missing_session"))
	}
	if len(b.networkRoster.Players) < 2 {
		return fmt.Errorf("%s", uiText(b.lang, "network.error.not_enough_players"))
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
		b.failNetwork(fmt.Sprintf(uiText(b.lang, "network.error.send_commands"), err.Error()))
	}
	st.submitted = true
	return &origTransition{next: b.networkWaitScreen()}
}

func mWithPlayer(m netplay.Message, player int) netplay.Message {
	m.Player = player
	return m
}

type networkWaitScreen struct {
	b      *sceneBuilder
	visual *netNextTurnScreen
	tick   int
}

func (b *sceneBuilder) networkWaitScreen() *networkWaitScreen {
	names := make([]string, len(b.networkRoster.Players))
	for i := range names {
		names[i] = fmt.Sprintf(uiText(b.lang, "network.player.fallback"), i+1)
	}
	for _, player := range b.networkRoster.Players {
		if player.ID < 0 || player.ID >= len(names) {
			continue
		}
		name := player.Name
		if name == "" {
			name = fmt.Sprintf(uiText(b.lang, "network.player.fallback"), player.ID+1)
		}
		names[player.ID] = name
	}
	var table *netplay.Table
	if b.networkTurn != nil {
		table = b.networkTurn.table
	}
	visual := b.netNextTurn(table, names, b.netMe)
	// 正式 update loop 是唯一封包 consumer；visual 只重用 renderer 與聊天輸入狀態。
	visual.sess = b.netSess
	return &networkWaitScreen{b: b, visual: visual}
}

func (s *networkWaitScreen) update(in shell.InputState) *origTransition {
	s.tick++
	st := s.b.networkTurn
	if st == nil {
		return nil
	}
	if s.visual != nil {
		s.visual.tick = s.tick
		s.visual.table = st.table
		if st.errorText == "" {
			s.visual.typeChatRunes(ebiten.AppendInputChars(nil))
			if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
				s.visual.backspaceChat()
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
				s.visual.sendChat()
			}
		}
	}
	if st.errorText == "" && s.b.netSess != nil {
		for _, m := range s.b.netSess.Poll() {
			s.handle(m)
		}
		if err := s.b.netSess.Err(); err != nil {
			s.b.failNetwork(fmt.Sprintf(uiText(s.b.lang, "network.error.connection"), err.Error()))
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
			s.b.failNetwork(uiText(s.b.lang, "network.error.commands_before_table"))
			return
		}
		if err := st.table.Add(m); err != nil {
			s.b.failNetwork(err.Error())
		}
	case netplay.KindTurnReady:
		if st.table == nil {
			s.b.failNetwork(uiText(s.b.lang, "network.error.ready_before_table"))
			return
		}
		if err := st.table.AddReady(m); err != nil {
			s.b.failNetwork(err.Error())
		}
	case netplay.KindDesync:
		s.b.failNetwork(m.Detail)
	case netplay.KindChat:
		if m.Text != "" && s.visual != nil {
			s.visual.chat.Append(m.Player, m.Text)
		}
	}
}

func (s *networkWaitScreen) replayCommands() error {
	st := s.b.networkTurn
	if st == nil || st.table == nil {
		return fmt.Errorf("%s", uiText(s.b.lang, "network.error.turn_data_missing"))
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
			return fmt.Errorf(uiText(s.b.lang, "network.error.replay_player"), p, err)
		}
	}
	if err := s.b.session.SetActiveSeat(0); err != nil {
		return err
	}
	hash := s.b.session.NetworkStateHash()
	if hash == "" {
		return fmt.Errorf("%s", uiText(s.b.lang, "network.error.fingerprint_missing"))
	}
	ready := netplay.Message{Kind: netplay.KindTurnReady, Turn: st.table.Turn(), StateHash: hash}
	if err := st.table.AddReady(mWithPlayer(ready, s.b.netMe)); err != nil {
		return err
	}
	if err := s.b.netSess.Send(ready); err != nil {
		return fmt.Errorf(uiText(s.b.lang, "network.error.send_fingerprint"), err)
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
		return fmt.Errorf("%s", uiText(b.lang, "network.error.turn_state_missing"))
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
		detail = uiText(b.lang, "network.error.stopped")
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
	st := s.b.networkTurn
	if s.visual == nil {
		dst.Fill(colorDarkNetwork)
		return
	}
	s.visual.tick = s.tick
	if st != nil {
		s.visual.table = st.table
	}
	s.visual.draw(dst)
	if st == nil || st.errorText == "" || s.b.fnt == nil {
		return
	}
	fillPanel(dst, 21, 360, 598, 88, color.RGBA{70, 16, 20, 245}, false)
	vector.StrokeRect(dst, 21, 360, 598, 88, 1, color.RGBA{230, 110, 110, 255}, false)
	networkWaitErrorTitleTextRect().drawCentered(dst, s.b.fnt,
		uiText(s.b.lang, "network.error.title"), 13, colorErrorNetwork)
	networkWaitErrorDetailTextRect().drawCentered(dst, s.b.fnt, st.errorText, 11, colorBodyNetwork)
	networkWaitErrorHintTextRect().drawCentered(dst, s.b.fnt,
		uiText(s.b.lang, "network.error.return_hint"), 11, colorBodyNetwork)
}

func networkWaitErrorTitleTextRect() textSafeRect {
	return textSafeRect{x: 29, y: 367, w: 582, h: 18, insetX: 4}
}

func networkWaitErrorDetailTextRect() textSafeRect {
	return textSafeRect{x: 29, y: 389, w: 582, h: 18, insetX: 4}
}

func networkWaitErrorHintTextRect() textSafeRect {
	return textSafeRect{x: 29, y: 416, w: 582, h: 18, insetX: 4}
}
