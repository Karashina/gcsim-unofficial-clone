package construct

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

type GeoConstructType int

const (
	GeoConstructInvalid GeoConstructType = iota
	GeoConstructNingSkill
	GeoConstructZhongliSkill
	GeoConstructTravellerSkill
	GeoConstructTravellerBurst
	GeoConstructAlbedoSkill
	GeoConstructIttoSkill
	GeoConstructKachinaSkill
	EndGeoConstructType
)

var ConstructString = [...]string{
	"Invalid",
	"NingSkill",
	"ZhongliSkill",
	"TravellerSkill",
	"TravellerBurst",
	"AlbedoSkill",
	"IttoSkill",
	"KachinaSkill",
}

var ConstructNameToKey = map[string]GeoConstructType{
	"ningguang":      GeoConstructNingSkill,
	"zhongli":        GeoConstructZhongliSkill,
	"traveler_skill": GeoConstructTravellerSkill,
	"traveler_burst": GeoConstructTravellerBurst,
	"albedo":         GeoConstructAlbedoSkill,
	"itto":           GeoConstructIttoSkill,
	"kachina":        GeoConstructKachinaSkill,
}

func (c GeoConstructType) String() string {
	return ConstructString[c]
}

type Construct interface {
	OnDestruct()
	Key() int
	Type() GeoConstructType
	Expiry() int
	IsLimited() bool
	Count() int
	Direction() geometry.Point
	Pos() geometry.Point
}

