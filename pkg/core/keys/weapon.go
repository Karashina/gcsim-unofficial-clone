package keys

import (
	"encoding/json"
	"errors"
	"strings"
)

type Weapon int

func (c *Weapon) MarshalJSON() ([]byte, error) {
	return json.Marshal(weaponNames[*c])
}

func (c *Weapon) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	s = strings.ToLower(s)
	for i, v := range weaponNames {
		if v == s {
			*c = Weapon(i)
			return nil
		}
	}
	return errors.New("unrecognized weapon key")
}

func (c Weapon) String() string {
	return weaponNames[c]
}

var weaponNames = []string{
	"",
	"absolution",
	"akuoumaru",
	"alleyhunter",
	"amenomakageuchi",
	"amosbow",
	"apprenticesnotes",
	"aquasimulacra",
	"aquilafavonia",
	"ashgravendrinkinghorn",
	"astralvulturescrimsonplumage",
	"athousandblazingsuns",
	"athousandfloatingdreams",
	"azurelight",
	"balladoftheboundlessblue",
	"balladofthefjords",
	"beaconofthereedsea",
	"beginnersprotector",
	"blackcliffagate",
	"blackclifflongsword",
	"blackcliffpole",
	"blackcliffslasher",
	"blackcliffwarbow",
	"blacktassel",
	"bloodtaintedgreatsword",
	"calamityofeshu",
	"calamityqueller",
	"cashflowsupervision",
	"chainbreaker",
	"cinnabarspindle",
	"cloudforged",
	"compoundbow",
	"coolsteel",
	"cranesechoingcall",
	"crescentpike",
	"crimsonmoonssemblance",
	"darkironsword",
	"deathmatch",
	"debateclub",
	"dialoguesofthedesertsages",
	"dodocotales",
	"dragonsbane",
	"dragonspinespear",
	"dullblade",
	"earthshaker",
	"elegyfortheend",
	"emeraldorb",
	"endoftheline",
	"engulfinglightning",
	"everlastingmoonglow",
	"eyeofperception",
	"fadingtwilight",
	"fangofthemountainking",
	"favoniuscodex",
	"favoniusgreatsword",
	"favoniuslance",
	"favoniussword",
	"favoniuswarbow",
	"ferrousshadow",
	"festeringdesire",
	"filletblade",
	"finaleofthedeep",
	"fleuvecendreferryman",
	"flowerwreathedfeathers",
	"fluteofezpitzal",
	"flowingpurity",
	"footprintoftherainbow",
	"forestregalia",
	"fracturedhalo",
	"freedomsworn",
	"frostbearer",
	"fruitfulhook",
	"fruitoffulfillment",
	"hakushinring",
	"halberd",
	"hamayumi",
	"harangeppakufutsu",
	"harbingerofdawn",
	"huntersbow",
	"hunterspath",
	"ibispiercer",
	"ironpoint",
	"ironsting",
	"jadefallssplendor",
	"kagotsurubeisshin",
	"kagurasverity",
	"katsuragikirinagamasa",
	"keyofkhajnisut",
	"kingssquire",
	"kitaincrossspear",
	"lightoffoliarincision",
	"lionsroar",
	"lithicblade",
	"lithicspear",
	"lostprayertothesacredwinds",
	"lumidouceelegy",
	"luxurioussealord",
	"magicguide",
	"mailedflower",
	"makhairaaquamarine",
	"mappamare",
	"memoryofdust",
	"messenger",
	"missivewindspear",
	"mistsplitterreforged",
	"mitternachtswaltz",
	"moonpiercer",
	"mountainbracingbolt",
	"mouunsmoon",
	"oathsworneye",
	"oldmercspal",
	"otherworldlystory",
	"peakpatrolsong",
	"pocketgrimoire",
	"polarstar",
	"portablepowersaw",
	"predator",
	"primordialjadecutter",
	"primordialjadewingedspear",
	"prospectorsdrill",
	"prototypeamber",
	"prototypearchaic",
	"prototypecrescent",
	"prototyperancour",
	"prototypestarglitter",
	"rainslasher",
	"rangegauge",
	"ravenbow",
	"recurvebow",
	"redhornstonethresher",
	"rightfulreward",
	"ringofyaxche",
	"royalbow",
	"royalgreatsword",
	"royalgrimoire",
	"royallongsword",
	"royalspear",
	"rust",
	"sacrificialbow",
	"sacrificialfragments",
	"sacrificialgreatsword",
	"sacrificialjade",
	"sacrificialsword",
	"sapwoodblade",
	"scionoftheblazingsun",
	"seasonedhuntersbow",
	"sequenceofsolitude",
	"serpentspine",
	"sharpshootersoath",
	"silvershowerheartstrings",
	"silversword",
	"skyridergreatsword",
	"skyridersword",
	"skywardatlas",
	"skywardblade",
	"skywardharp",
	"skywardpride",
	"skywardspine",
	"slingshot",
	"snowtombedstarsilver",
	"solarpearl",
	"songofbrokenpines",
	"songofstillness",
	"splendoroftranquilwaters",
	"staffofhoma",
	"staffofthescarletsands",
	"starcallerswatch",
	"sturdybone",
	"summitshaper",
	"sunnymorningsleepin",
	"surfsup",
	"swordofdescension",
	"swordofnarzissenkreuz",
	"symphonistofscents",
	"talkingstick",
	"tamayurateinoohanashi",
	"thealleyflash",
	"thebell",
	"theblacksword",
	"thecatch",
	"thedockhandsassistant",
	"thefirstgreatmagic",
	"theflute",
	"thestringless",
	"theunforged",
	"theviridescenthunt",
	"thewidsith",
	"thrillingtalesofdragonslayers",
	"thunderingpulse",
	"tidalshadow",
	"tomeoftheeternalflow",
	"toukaboushigure",
	"travelershandysword",
	"tulaytullahsremembrance",
	"twinnephrite",
	"ultimateoverlordsmegamagicsword",
	"urakumisugiri",
	"verdict",
	"vividnotions",
	"vortexvanquisher",
	"wanderingevenstar",
	"wastergreatsword",
	"wavebreakersfin",
	"waveridingwhirl",
	"whiteblind",
	"whiteirongreatsword",
	"whitetassel",
	"windblumeode",
	"wineandsong",
	"wolffang",
	"wolfsgravestone",
	"xiphosmoonlight",
}

const (
	NoWeapon Weapon = iota
	Absolution
	Akuoumaru
	AlleyHunter
	AmenomaKageuchi
	AmosBow
	ApprenticesNotes
	AquaSimulacra
	AquilaFavonia
	AshGravenDrinkingHorn
	AstralVulturesCrimsonPlumage
	AzureLight
	AThousandBlazingSuns
	AThousandFloatingDreams
	BalladOfTheBoundlessBlue
	BalladOfTheFjords
	BeaconOfTheReedSea
	BeginnersProtector
	BlackcliffAgate
	BlackcliffLongsword
	BlackcliffPole
	BlackcliffSlasher
	BlackcliffWarbow
	BlackTassel
	BloodtaintedGreatsword
	CalamityOfEshu
	CalamityQueller
	CashflowSupervision
	ChainBreaker
	CinnabarSpindle
	Cloudforged
	CompoundBow
	CoolSteel
	CranesEchoingCall
	CrescentPike
	CrimsonMoonsSemblance
	DarkIronSword
	Deathmatch
	DebateClub
	DialoguesOfTheDesertSages
	DodocoTales
	DragonsBane
	DragonspineSpear
	DullBlade
	EarthShaker
	ElegyForTheEnd
	EmeraldOrb
	EndOfTheLine
	EngulfingLightning
	EverlastingMoonglow
	EyeOfPerception
	FadingTwilight
	FangOfTheMountainKing
	FavoniusCodex
	FavoniusGreatsword
	FavoniusLance
	FavoniusSword
	FavoniusWarbow
	FerrousShadow
	FesteringDesire
	FilletBlade
	FinaleOfTheDeep
	FleuveCendreFerryman
	FlowerWreathedFeathers
	FluteOfEzpitzal
	FlowingPurity
	FootprintOfTheRainbow
	ForestRegalia
	FracturedHalo
	FreedomSworn
	Frostbearer
	FruitfulHook
	FruitOfFulfillment
	HakushinRing
	Halberd
	Hamayumi
	HaranGeppakuFutsu
	HarbingerOfDawn
	HuntersBow
	HuntersPath
	IbisPiercer
	IronPoint
	IronSting
	JadefallsSplendor
	KagotsurubeIsshin
	KagurasVerity
	KatsuragikiriNagamasa
	KeyOfKhajNisut
	KingsSquire
	KitainCrossSpear
	LightOfFoliarIncision
	LionsRoar
	LithicBlade
	LithicSpear
	LostPrayerToTheSacredWinds
	LumidouceElegy
	LuxuriousSeaLord
	MagicGuide
	MailedFlower
	MakhairaAquamarine
	MappaMare
	MemoryOfDust
	Messenger
	MissiveWindspear
	MistsplitterReforged
	MitternachtsWaltz
	Moonpiercer
	MountainBracingBolt
	MouunsMoon
	OathswornEye
	OldMercsPal
	OtherworldlyStory
	PeakPatrolSong
	PocketGrimoire
	PolarStar
	PortablePowerSaw
	Predator
	PrimordialJadeCutter
	PrimordialJadeWingedSpear
	ProspectorsDrill
	PrototypeAmber
	PrototypeArchaic
	PrototypeCrescent
	PrototypeRancour
	PrototypeStarglitter
	Rainslasher
	RangeGauge
	RavenBow
	RecurveBow
	RedhornStonethresher
	RightfulReward
	RingOfYaxche
	RoyalBow
	RoyalGreatsword
	RoyalGrimoire
	RoyalLongsword
	RoyalSpear
	Rust
	SacrificialBow
	SacrificialFragments
	SacrificialGreatsword
	SacrificialJade
	SacrificialSword
	SapwoodBlade
	ScionOfTheBlazingSun
	SeasonedHuntersBow
	SequenceOfSolitude
	SerpentSpine
	SharpshootersOath
	SilvershowerHeartstrings
	SilverSword
	SkyriderGreatsword
	SkyriderSword
	SkywardAtlas
	SkywardBlade
	SkywardHarp
	SkywardPride
	SkywardSpine
	Slingshot
	SnowTombedStarsilver
	SolarPearl
	SongOfBrokenPines
	SongOfStillness
	SplendorOfTranquilWaters
	StaffOfHoma
	StaffOfTheScarletSands
	StarcallersWatch
	SturdyBone
	SummitShaper
	SunnyMorningSleepIn
	SurfsUp
	SwordOfDescension
	SwordOfNarzissenkreuz
	SymphonistOfScents
	TalkingStick
	TamayurateiNoOhanashi
	TheAlleyFlash
	TheBell
	TheBlackSword
	TheCatch
	TheDockhandsAssistant
	TheFirstGreatMagic
	TheFlute
	TheStringless
	TheUnforged
	TheViridescentHunt
	TheWidsith
	ThrillingTalesOfDragonSlayers
	ThunderingPulse
	TidalShadow
	TomeOfTheEternalFlow
	ToukabouShigure
	TravelersHandySword
	TulaytullahsRemembrance
	TwinNephrite
	UltimateOverlordsMegaMagicSword
	UrakuMisugiri
	Verdict
	VividNotions
	VortexVanquisher
	WanderingEvenstar
	WasterGreatsword
	WavebreakersFin
	WaveridingWhirl
	Whiteblind
	WhiteIronGreatsword
	WhiteTassel
	WindblumeOde
	WineAndSong
	WolfFang
	WolfsGravestone
	XiphosMoonlight
)
