package target

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
)

func (t *Target) WillApplyEle(tag attacks.ICDTag, grp attacks.ICDGroup, char int) float64 {
	// タグなしの場合はICDなし
	if tag == attacks.ICDTagNone {
		return 1
	}
	// タイマーの開始が必要か確認
	x := t.icdTagOnTimer[char][tag]
	if !t.icdTagOnTimer[char][tag] {
		t.icdTagOnTimer[char][tag] = true
		t.ResetTagCounterAfterDelay(tag, grp, char)
	}
	val := t.icdTagCounter[char][tag]
	t.icdTagCounter[char][tag]++
	// カウンターが長さを超える場合はグループ序列に0を使用
	groupSeq := attacks.ICDGroupEleApplicationSequence[grp][len(attacks.ICDGroupEleApplicationSequence[grp])-1]
	if val < len(attacks.ICDGroupEleApplicationSequence[grp]) {
		groupSeq = attacks.ICDGroupEleApplicationSequence[grp][val]
	}
	t.Core.Log.NewEvent("ele icd check", glog.LogICDEvent, char).
		Write("grp", grp).
		Write("target", t.key).
		Write("tag", tag).
		Write("counter", val).
		Write("val", groupSeq).
		Write("group on timer", x)
	return groupSeq
}
func (t *Target) GroupTagDamageMult(tag attacks.ICDTag, grp attacks.ICDGroup, char int) float64 {
	// タイマーの開始が必要か確認
	if !t.icdDamageTagOnTimer[char][tag] {
		t.icdDamageTagOnTimer[char][tag] = true
		t.ResetDamageCounterAfterDelay(tag, grp, char)
	}
	val := t.icdDamageTagCounter[char][tag]
	t.icdDamageTagCounter[char][tag]++
	// カウンターが長さを超える場合はグループ序列に0を使用
	groupSeq := attacks.ICDGroupDamageSequence[grp][len(attacks.ICDGroupDamageSequence[grp])-1]
	if val < len(attacks.ICDGroupDamageSequence[grp]) {
		groupSeq = attacks.ICDGroupDamageSequence[grp][val]
	}
	return groupSeq
}
func (t *Target) ResetDamageCounterAfterDelay(tag attacks.ICDTag, grp attacks.ICDGroup, char int) {
	t.Core.Tasks.Add(func() {
		// カウンターを0にリセット
		t.icdDamageTagCounter[char][tag] = 0
		t.icdDamageTagOnTimer[char][tag] = false
		t.Core.Log.NewEvent("damage counter reset", glog.LogICDEvent, char).
			Write("tag", tag).
			Write("grp", grp)
	}, attacks.ICDGroupResetTimer[grp]-1)
	t.Core.Log.NewEvent("damage reset timer set", glog.LogICDEvent, char).
		Write("tag", tag).
		Write("grp", grp).
		Write("reset", t.Core.F+attacks.ICDGroupResetTimer[grp]-1)
}
func (t *Target) ResetTagCounterAfterDelay(tag attacks.ICDTag, grp attacks.ICDGroup, char int) {
	t.Core.Tasks.Add(func() {
		// カウンターを0にリセット
		t.icdTagCounter[char][tag] = 0
		t.icdTagOnTimer[char][tag] = false
		t.Core.Log.NewEvent("ele app counter reset", glog.LogICDEvent, char).
			Write("tag", tag).
			Write("grp", grp)
	}, attacks.ICDGroupResetTimer[grp]-1)
	t.Core.Log.NewEvent("ele app reset timer set", glog.LogICDEvent, char).
		Write("tag", tag).
		Write("grp", grp).
		Write("reset", t.Core.F+attacks.ICDGroupResetTimer[grp]-1)
}
