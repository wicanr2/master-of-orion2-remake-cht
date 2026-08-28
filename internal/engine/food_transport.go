package engine

// OriginalFoodTransportResult 對映 sub_DF8F0 的玩家可見帝國輸出。
type OriginalFoodTransportResult struct {
	FoodFreighted     int
	SurplusFreighters int
	TotalSurplusFood  int
	TotalDeficitFood  int
}

// OriginalFoodTransport 對映 sub_DF8F0 的食物／貨運艦彙總主形狀。
// 一艘貨運艦搬運一單位食物；一個在途 settler 先占用五艘。封鎖殖民地不參與
// 跨星運輸。ColonyOutput 使用半食物單位，因此出口向下取整、缺口向上取整。
func OriginalFoodTransport(ps PlayerState, colonies []ColonyState, blockaded []bool) (OriginalFoodTransportResult, bool) {
	if ps.ActiveFreighters < 0 || ps.SettlersFreighted < 0 ||
		(blockaded != nil && len(blockaded) != len(colonies)) {
		return OriginalFoodTransportResult{}, false
	}
	out := OriginalFoodTransportResult{}
	for i := range colonies {
		if blockaded != nil && blockaded[i] {
			continue
		}
		surplusHalf := RunColonyTurn(colonies[i]).FoodSurplusHalf
		if surplusHalf >= 0 {
			out.TotalSurplusFood += surplusHalf / 2
		} else {
			out.TotalDeficitFood += (-surplusHalf + 1) / 2
		}
	}
	available := ps.ActiveFreighters - 5*ps.SettlersFreighted
	transferable := out.TotalSurplusFood
	if out.TotalDeficitFood < transferable {
		transferable = out.TotalDeficitFood
	}
	if available > 0 {
		out.FoodFreighted = transferable
		if available < out.FoodFreighted {
			out.FoodFreighted = available
		}
	}
	remaining := available - out.FoodFreighted
	if remaining != 0 {
		out.SurplusFreighters = remaining
	} else {
		out.SurplusFreighters = -(transferable - out.FoodFreighted)
	}
	return out, true
}
