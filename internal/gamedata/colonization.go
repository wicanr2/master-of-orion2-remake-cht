package gamedata

// ColonizationStartsWithWorker 對應 sub_E5EB3 @ 0xE5F74..0xE5FA7。
// 原版在行星自然食物為零、Lithovore 或 Cybernetic 任一成立時，將第一位殖民者
// 的 packed job 設為 1（工人）；否則清成 0（農夫）。
func ColonizationStartsWithWorker(naturalFood int, lithovore, cybernetic bool) bool {
	return naturalFood <= 0 || lithovore || cybernetic
}
