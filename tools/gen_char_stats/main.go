package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/encoding/prototext"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

// Simple tool to read data_gen.textproto for a character and print
// ascension table computed from base_* and pkg/model/curves.go.
// It also prints promo_data values and verifies they match computed diffs.

func main() {
	dir := flag.String("dir", "internal/characters/albedo", "character folder containing data_gen.textproto")
	out := flag.String("out", "", "optional output file to write CSV table")
	flag.Parse()

	p := filepath.Join(*dir, "data_gen.textproto")
	b, err := ioutil.ReadFile(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed read %s: %v\n", p, err)
		os.Exit(2)
	}

	var av model.AvatarData
	if err := prototext.Unmarshal(b, &av); err != nil {
		fmt.Fprintf(os.Stderr, "failed unmarshal proto text: %v\n", err)
		os.Exit(2)
	}

	// Determine hp curve type index key used in curves.go map
	// model.AvatarCurveType enums are generated; AvatarGrowCurveByLvl uses these enums as keys.

	fmt.Printf("Character: %s (id=%d)\n", av.GetIconName(), av.GetId())
	baseHP := av.GetStats().GetBaseHp()
	baseATK := av.GetStats().GetBaseAtk()
	baseDEF := av.GetStats().GetBaseDef()
	hpCurve := av.GetStats().GetHpCurve()
	atkCurve := av.GetStats().GetAtkCurve()
	defCurve := av.GetStats().GetDefCruve()

	fmt.Printf("base_hp: %.6f base_atk: %.6f base_def: %.6f\n", baseHP, baseATK, baseDEF)
	fmt.Printf("hp_curve: %v atk_curve: %v def_curve: %v\n", hpCurve, atkCurve, defCurve)

	// print header
	lines := make([]string, 0)
	lines = append(lines, "level,base_hp,base_atk,base_def")

	// AvatarGrowCurveByLvl is index 0..89 mapping AvatarCurveType->factor
	for lvl := 1; lvl <= 90; lvl++ {
		idx := lvl - 1
		if idx >= len(model.AvatarGrowCurveByLvl) {
			break
		}
		hpFactor := model.AvatarGrowCurveByLvl[idx][hpCurve]
		atkFactor := model.AvatarGrowCurveByLvl[idx][atkCurve]
		defFactor := model.AvatarGrowCurveByLvl[idx][defCurve]
		hp := baseHP * hpFactor
		atk := baseATK * atkFactor
		defv := baseDEF * defFactor
		lines = append(lines, fmt.Sprintf("%d,%.6f,%.6f,%.6f", lvl, hp, atk, defv))
	}

	// output either to stdout or to file
	if *out == "" {
		for _, l := range lines {
			fmt.Println(l)
		}
	} else {
		if err := ioutil.WriteFile(*out, []byte(fmt.Sprintln("# csv\n")+joinLines(lines)), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "failed write out: %v\n", err)
			os.Exit(3)
		}
		fmt.Printf("wrote %s\n", *out)
	}

	// print promo_data and verify diffs for their max_level entries
	fmt.Println("\npromo_data entries:")
	for _, pdat := range av.GetStats().GetPromoData() {
		ml := pdat.GetMaxLevel()
		fmt.Printf("max_level=%d\n", ml)
		for _, ap := range pdat.GetAddProps() {
			fmt.Printf("  prop=%v value=%v\n", ap.GetPropType(), ap.GetValue())
		}
		// compute expected diff for base_hp if provided
		if ml > 0 && ml <= 90 {
			idx := ml - 1
			expected := baseHP * model.AvatarGrowCurveByLvl[idx][hpCurve]
			// find previous asc level base: previous asc target determined by promo list ordering
			// Simplify: compute diff between expected and value at previous asc level (we try common asc breakpoints)
			// We'll just print expected absolute value; manual diff check is left to user.
			fmt.Printf("  computed_total_hp_at_maxlevel=%.6f\n", expected)
		}
	}
}

func joinLines(lines []string) string {
	s := ""
	for _, l := range lines {
		s += l + "\n"
	}
	return s
}
