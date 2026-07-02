// Code generated from openorion2 gamestate.h enums. DO NOT EDIT by hand.
// 由 scripts/gen-enums(scratchpad genenum.py)自 openorion2 gamestate.h 生成。
package gamedata

// MultiplayerType
type MultiplayerType int
const (
	Single MultiplayerType = 0
	Hotseat MultiplayerType = 1
	Network MultiplayerType = 2
)

// PlanetType
type PlanetType int
const (
	ASTEROIDS PlanetType = 1
	GAS_GIANT PlanetType = 2
	HABITABLE PlanetType = 3
)

// PlanetSize
type PlanetSize int
const (
	TINY_PLANET PlanetSize = 0
	SMALL_PLANET PlanetSize = 1
	MEDIUM_PLANET PlanetSize = 2
	LARGE_PLANET PlanetSize = 3
	HUGE_PLANET PlanetSize = 4
)

// PlanetGravity
type PlanetGravity int
const (
	LOW_G PlanetGravity = 0
	NORMAL_G PlanetGravity = 1
	HEAVY_G PlanetGravity = 2
)

// PlanetClimate
type PlanetClimate int
const (
	TOXIC PlanetClimate = 0
	RADIATED PlanetClimate = 1
	BARREN PlanetClimate = 2
	DESERT PlanetClimate = 3
	TUNDRA PlanetClimate = 4
	OCEAN PlanetClimate = 5
	SWAMP PlanetClimate = 6
	ARID PlanetClimate = 7
	TERRAN PlanetClimate = 8
	GAIA PlanetClimate = 9
)

// PlanetMinerals
type PlanetMinerals int
const (
	ULTRA_POOR PlanetMinerals = 0
	POOR PlanetMinerals = 1
	ABUNDANT PlanetMinerals = 2
	RICH PlanetMinerals = 3
	ULTRA_RICH PlanetMinerals = 4
)

// ColonistRace
type ColonistRace int
const (
	ANDROID ColonistRace = 8
	NATIVE ColonistRace = 9
)

// ColonistFlags
type ColonistFlags int
const (
	WORKING ColonistFlags = 1
	PRISONER ColonistFlags = 2
)

// ColonistJob
type ColonistJob int
const (
	FARMER ColonistJob = 0
	WORKER ColonistJob = 1
	SCIENTIST ColonistJob = 2
)

// StarSize
type StarSize int
const (
	Large StarSize = 0
	Medium StarSize = 1
	Small StarSize = 2
	Tiny StarSize = 3
)

// SpectralClass
type SpectralClass int
const (
	Blue SpectralClass = 0
	White SpectralClass = 1
	Yellow SpectralClass = 2
	Orange SpectralClass = 3
	Red SpectralClass = 4
	Brown SpectralClass = 5
	BlackHole SpectralClass = 6
)

// StarKnowledge
type StarKnowledge int
const (
	STAR_UNEXPLORED StarKnowledge = 0
	STAR_NAME_ONLY StarKnowledge = 1
	STAR_CHARTED StarKnowledge = 2
	STAR_VISITED StarKnowledge = 3
)

// SpecialType
type SpecialType int
const (
	NO_SPECIAL SpecialType = 0
	BAD_SPECIAL1 SpecialType = 1
	SPACE_DEBRIS SpecialType = 2
	PIRATE_CACHE SpecialType = 3
	GOLD_DEPOSITS SpecialType = 4
	GEM_DEPOSITS SpecialType = 5
	NATIVES SpecialType = 6
	SPLINTER_COLONY SpecialType = 7
	LOST_HERO SpecialType = 8
	BAD_SPECIAL2 SpecialType = 9
	ANCIENT_ARTIFACTS SpecialType = 10
	ORION_SPECIAL SpecialType = 11
)

// ResearchStatus
type ResearchStatus int
const (
	RSTATE_DISABLED ResearchStatus = 0
	RSTATE_RESEARCHABLE ResearchStatus = 1
	RSTATE_READY ResearchStatus = 2
	RSTATE_KNOWN ResearchStatus = 3
)

// ResearchArea
type ResearchArea int
const (
	RESEARCH_BIOLOGY ResearchArea = 0
	RESEARCH_POWER ResearchArea = 1
	RESEARCH_PHYSICS ResearchArea = 2
	RESEARCH_CONSTRUCTION ResearchArea = 3
	RESEARCH_FIELDS ResearchArea = 4
	RESEARCH_CHEMISTRY ResearchArea = 5
	RESEARCH_COMPUTERS ResearchArea = 6
	RESEARCH_SOCIOLOGY ResearchArea = 7
)

// ResearchTopic
type ResearchTopic int
const (
	TOPIC_STARTING_TECH ResearchTopic = 0
	TOPIC_ADVANCED_BIOLOGY ResearchTopic = 1
	TOPIC_ADVANCED_CHEMISTRY ResearchTopic = 2
	TOPIC_ADVANCED_CONSTRUCTION ResearchTopic = 3
	TOPIC_ADVANCED_ENGINEERING ResearchTopic = 4
	TOPIC_ADVANCED_FUSION ResearchTopic = 5
	TOPIC_ADVANCED_GOVERNMENTS ResearchTopic = 6
	TOPIC_ADVANCED_MAGNETISM ResearchTopic = 7
	TOPIC_ADVANCED_MANUFACTURING ResearchTopic = 8
	TOPIC_ADVANCED_METALLURGY ResearchTopic = 9
	TOPIC_MILITARY_TACTICS ResearchTopic = 10
	TOPIC_ADVANCED_ROBOTICS ResearchTopic = 11
	TOPIC_TEACHING_METHODS ResearchTopic = 12
	TOPIC_ANTIMATTER_FISSION ResearchTopic = 13
	TOPIC_ARTIFICIAL_CONSCIOUSNESS ResearchTopic = 14
	TOPIC_ARTIFICIAL_INTELLIGENCE ResearchTopic = 15
	TOPIC_ARTIFICIAL_GRAVITY ResearchTopic = 16
	TOPIC_ARTIFICIAL_LIFE ResearchTopic = 17
	TOPIC_ASTRO_BIOLOGY ResearchTopic = 18
	TOPIC_ASTRO_CONSTRUCTION ResearchTopic = 19
	TOPIC_ASTRO_ENGINEERING ResearchTopic = 20
	TOPIC_CAPSULE_CONSTRUCTION ResearchTopic = 21
	TOPIC_CHEMISTRY ResearchTopic = 22
	TOPIC_COLD_FUSION ResearchTopic = 23
	TOPIC_CYBERTECHNICS ResearchTopic = 24
	TOPIC_CYBERTRONICS ResearchTopic = 25
	TOPIC_DISTORTION_FIELDS ResearchTopic = 26
	TOPIC_ELECTROMAGNETIC_REFRACTION ResearchTopic = 27
	TOPIC_ELECTRONICS ResearchTopic = 28
	TOPIC_ENGINEERING ResearchTopic = 29
	TOPIC_EVOLUTIONARY_GENETICS ResearchTopic = 30
	TOPIC_FUSION_PHYSICS ResearchTopic = 31
	TOPIC_GALACTIC_ECONOMICS ResearchTopic = 32
	TOPIC_GALACTIC_NETWORKING ResearchTopic = 33
	TOPIC_GENETIC_ENGINEERING ResearchTopic = 34
	TOPIC_GENETIC_MUTATIONS ResearchTopic = 35
	TOPIC_GRAVITIC_FIELDS ResearchTopic = 36
	TOPIC_HIGH_ENERGY_DISTRIBUTION ResearchTopic = 37
	TOPIC_HYPER_DIMENSIONAL_FISSION ResearchTopic = 38
	TOPIC_HYPER_DIMENSIONAL_PHYSICS ResearchTopic = 39
	TOPIC_INTERPHASED_FISSION ResearchTopic = 40
	TOPIC_ION_FISSION ResearchTopic = 41
	TOPIC_SUPERSCALAR_CONSTRUCTION ResearchTopic = 42
	TOPIC_MACRO_ECONOMICS ResearchTopic = 43
	TOPIC_MACRO_GENETICS ResearchTopic = 44
	TOPIC_MAGNETO_GRAVITICS ResearchTopic = 45
	TOPIC_MATTER_ENERGY_CONVERSION ResearchTopic = 46
	TOPIC_MOLECULAR_COMPRESSION ResearchTopic = 47
	TOPIC_MOLECULAR_CONTROL ResearchTopic = 48
	TOPIC_MOLECULATRONICS ResearchTopic = 49
	TOPIC_MOLECULAR_MANIPULATION ResearchTopic = 50
	TOPIC_MULTIDIMENSIONAL_PHYSICS ResearchTopic = 51
	TOPIC_MULTIPHASED_PHYSICS ResearchTopic = 52
	TOPIC_NANO_TECHNOLOGY ResearchTopic = 53
	TOPIC_NEUTRINO_PHYSICS ResearchTopic = 54
	TOPIC_NUCLEAR_FISSION ResearchTopic = 55
	TOPIC_OPTRONICS ResearchTopic = 56
	TOPIC_PHYSICS ResearchTopic = 57
	TOPIC_PLANETOID_CONSTRUCTION ResearchTopic = 58
	TOPIC_PLASMA_PHYSICS ResearchTopic = 59
	TOPIC_POSITRONICS ResearchTopic = 60
	TOPIC_QUANTUM_FIELDS ResearchTopic = 61
	TOPIC_ROBOTICS ResearchTopic = 62
	TOPIC_SERVO_MECHANICS ResearchTopic = 63
	TOPIC_SUBSPACE_FIELDS ResearchTopic = 64
	TOPIC_SUBSPACE_PHYSICS ResearchTopic = 65
	TOPIC_TACHYON_PHYSICS ResearchTopic = 66
	TOPIC_TECTONIC_ENGINEERING ResearchTopic = 67
	TOPIC_TEMPORAL_FIELDS ResearchTopic = 68
	TOPIC_TEMPORAL_PHYSICS ResearchTopic = 69
	TOPIC_TRANS_GENETICS ResearchTopic = 70
	TOPIC_TRANSWARP_FIELDS ResearchTopic = 71
	TOPIC_WARP_FIELDS ResearchTopic = 72
	TOPIC_XENO_RELATIONS ResearchTopic = 73
	TOPIC_XENON_TECHNOLOGY ResearchTopic = 74
	TOPIC_HYPER_BIOLOGY ResearchTopic = 75
	TOPIC_HYPER_POWER ResearchTopic = 76
	TOPIC_HYPER_PHYSICS ResearchTopic = 77
	TOPIC_HYPER_CONSTRUCTION ResearchTopic = 78
	TOPIC_HYPER_FIELDS ResearchTopic = 79
	TOPIC_HYPER_CHEMISTRY ResearchTopic = 80
	TOPIC_HYPER_COMPUTERS ResearchTopic = 81
	TOPIC_HYPER_SOCIOLOGY ResearchTopic = 82
)

// Technology
type Technology int
const (
	TECH_NONE Technology = 0
	TECH_ACHILLES_TARGETING_UNIT Technology = 1
	TECH_ADAMANTIUM_ARMOR Technology = 2
	TECH_ADVANCED_CITY_PLANNING Technology = 3
	TECH_ADVANCED_DAMAGE_CONTROL Technology = 4
	TECH_ALIEN_MANAGEMENT_CENTER Technology = 5
	TECH_ANDROID_FARMERS Technology = 6
	TECH_ANDROID_SCIENTISTS Technology = 7
	TECH_ANDROID_WORKERS Technology = 8
	TECH_ANTIGRAV_HARNESS Technology = 9
	TECH_ANTIMATTER_BOMB Technology = 10
	TECH_ANTIMATTER_DRIVE Technology = 11
	TECH_ANTIMATTER_TORPEDOES Technology = 12
	TECH_ANTIMISSILE_ROCKETS Technology = 13
	TECH_ARMOR_BARRACKS Technology = 14
	TECH_ARTEMIS_SYSTEM_NET Technology = 15
	TECH_PLANET_CONSTRUCTION Technology = 16
	TECH_ASSAULT_SHUTTLES Technology = 17
	TECH_ASTRO_UNIVERSITY Technology = 18
	TECH_ATMOSPHERIC_RENEWER Technology = 19
	TECH_AUGMENTED_ENGINES Technology = 20
	TECH_AUTOLAB Technology = 21
	TECH_AUTOMATED_FACTORIES Technology = 22
	TECH_AUTOMATED_REPAIR_UNIT Technology = 23
	TECH_BATTLEOIDS Technology = 24
	TECH_BATTLE_PODS Technology = 25
	TECH_BATTLE_SCANNER Technology = 26
	TECH_BATTLESTATION Technology = 27
	TECH_BIOTERMINATOR Technology = 28
	TECH_BIOMORPHIC_FUNGI Technology = 29
	TECH_BLACK_HOLE_GENERATOR Technology = 30
	TECH_BOMBER_BAYS Technology = 31
	TECH_CAPITOL Technology = 32
	TECH_CLASS_I_SHIELD Technology = 33
	TECH_CLASS_III_SHIELD Technology = 34
	TECH_CLASS_V_SHIELD Technology = 35
	TECH_CLASS_VII_SHIELD Technology = 36
	TECH_CLASS_X_SHIELD Technology = 37
	TECH_CLOAKING_DEVICE Technology = 38
	TECH_CLONING_CENTER Technology = 39
	TECH_COLONY_BASE Technology = 40
	TECH_COLONY_SHIP Technology = 41
	TECH_CONFEDERATION Technology = 42
	TECH_CYBERSECURITY_LINK Technology = 43
	TECH_CYBERTRONIC_COMPUTER Technology = 44
	TECH_DAMPER_FIELD Technology = 45
	TECH_DAUNTLESS_GUIDANCE_SYSTEM Technology = 46
	TECH_DEATH_RAY Technology = 47
	TECH_DEATH_SPORES Technology = 48
	TECH_DEEP_CORE_MINING Technology = 49
	TECH_CORE_WASTE_DUMPS Technology = 50
	TECH_DEUTERIUM_FUEL_CELLS Technology = 51
	TECH_DIMENSIONAL_PORTAL Technology = 52
	TECH_DISPLACEMENT_DEVICE Technology = 53
	TECH_DISRUPTER_CANNON Technology = 54
	TECH_DOOM_STAR_CONSTRUCTION Technology = 55
	TECH_REINFORCED_HULL Technology = 56
	TECH_ECM_JAMMER Technology = 57
	TECH_ELECTRONIC_COMPUTER Technology = 58
	TECH_EMISSIONS_GUIDANCE_SYSTEM Technology = 59
	TECH_ENERGY_ABSORBER Technology = 60
	TECH_BIOSPHERES Technology = 61
	TECH_EVOLUTIONARY_MUTATION Technology = 62
	TECH_EXTENDED_FUEL_TANKS Technology = 63
	TECH_FAST_MISSILE_RACKS Technology = 64
	TECH_FEDERATION Technology = 65
	TECH_FIGHTER_BAYS Technology = 66
	TECH_FIGHTER_GARRISON Technology = 67
	TECH_FOOD_REPLICATORS Technology = 68
	TECH_FREIGHTERS Technology = 69
	TECH_FUSION_BEAM Technology = 70
	TECH_FUSION_BOMB Technology = 71
	TECH_FUSION_DRIVE Technology = 72
	TECH_FUSION_RIFLE Technology = 73
	TECH_GAIA_TRANSFORMATION Technology = 74
	TECH_GALACTIC_CURRENCY_EXCHANGE Technology = 75
	TECH_GALACTIC_CYBERNET Technology = 76
	TECH_GALACTIC_UNIFICATION Technology = 77
	TECH_GAUSS_CANNON Technology = 78
	TECH_GRAVITON_BEAM Technology = 79
	TECH_GYRO_DESTABILIZER Technology = 80
	TECH_HARD_SHIELDS Technology = 81
	TECH_HEAVY_ARMOR Technology = 82
	TECH_HEAVY_FIGHTER_BAYS Technology = 83
	TECH_HEIGHTENED_INTELLIGENCE Technology = 84
	TECH_HIGH_ENERGY_FOCUS Technology = 85
	TECH_HOLO_SIMULATOR Technology = 86
	TECH_HYDROPONIC_FARM Technology = 87
	TECH_HYPER_DRIVE Technology = 88
	TECH_MEGAFLUXERS Technology = 89
	TECH_HYPERX_CAPACITORS Technology = 90
	TECH_HYPERSPACE_COMMUNICATIONS Technology = 91
	TECH_IMPERIUM Technology = 92
	TECH_INERTIAL_NULLIFIER Technology = 93
	TECH_INERTIAL_STABILIZER Technology = 94
	TECH_INTERPHASED_DRIVE Technology = 95
	TECH_ION_DRIVE Technology = 96
	TECH_ION_PULSE_CANNON Technology = 97
	TECH_IRIDIUM_FUEL_CELLS Technology = 98
	TECH_JUMP_GATE Technology = 99
	TECH_LASER_CANNON Technology = 100
	TECH_LASER_RIFLE Technology = 101
	TECH_LIGHTNING_FIELD Technology = 102
	TECH_MARINE_BARRACKS Technology = 103
	TECH_MASS_DRIVER Technology = 104
	TECH_MAULER_DEVICE Technology = 105
	TECH_MERCULITE_MISSILE Technology = 106
	TECH_MICROBIOTICS Technology = 107
	TECH_MICROLITE_CONSTRUCTION Technology = 108
	TECH_OUTPOST_SHIP Technology = 109
	TECH_MOLECULARTRONIC_COMPUTER Technology = 110
	TECH_MULTIWAVE_ECM_JAMMER Technology = 111
	TECH_MULTIPHASED_SHIELDS Technology = 112
	TECH_NANO_DISASSEMBLERS Technology = 113
	TECH_NEURAL_SCANNER Technology = 114
	TECH_NEUTRON_BLASTER Technology = 115
	TECH_NEUTRON_SCANNER Technology = 116
	TECH_NEUTRONIUM_ARMOR Technology = 117
	TECH_NEUTRONIUM_BOMB Technology = 118
	TECH_NUCLEAR_BOMB Technology = 119
	TECH_NUCLEAR_DRIVE Technology = 120
	TECH_NUCLEAR_MISSILE Technology = 121
	TECH_OPTRONIC_COMPUTER Technology = 122
	TECH_PARTICLE_BEAM Technology = 123
	TECH_PERSONAL_SHIELD Technology = 124
	TECH_PHASE_SHIFTER Technology = 125
	TECH_PHASING_CLOAK Technology = 126
	TECH_PHASOR Technology = 127
	TECH_PHASOR_RIFLE Technology = 128
	TECH_PLANETARY_BARRIER_SHIELD Technology = 129
	TECH_PLANETARY_FLUX_SHIELD Technology = 130
	TECH_PLANETARY_GRAVITY_GENERATOR Technology = 131
	TECH_PLANETARY_MISSILE_BASE Technology = 132
	TECH_GROUND_BATTERIES Technology = 133
	TECH_PLANETARY_RADIATION_SHIELD Technology = 134
	TECH_PLANETARY_STOCK_EXCHANGE Technology = 135
	TECH_PLANETARY_SUPERCOMPUTER Technology = 136
	TECH_PLASMA_CANNON Technology = 137
	TECH_PLASMA_RIFLE Technology = 138
	TECH_PLASMA_TORPEDOES Technology = 139
	TECH_PLASMA_WEB Technology = 140
	TECH_PLEASURE_DOME Technology = 141
	TECH_POLLUTION_PROCESSOR Technology = 142
	TECH_POSITRONIC_COMPUTER Technology = 143
	TECH_POWERED_ARMOR Technology = 144
	TECH_PULSE_RIFLE Technology = 145
	TECH_PROTON_TORPEDOES Technology = 146
	TECH_PSIONICS Technology = 147
	TECH_PULSAR Technology = 148
	TECH_PULSON_MISSILE Technology = 149
	TECH_QUANTUM_DETONATOR Technology = 150
	TECH_RANGEMASTER_UNIT Technology = 151
	TECH_RECYCLOTRON Technology = 152
	TECH_REFLECTION_FIELD Technology = 153
	TECH_ROBOTIC_FACTORY Technology = 154
	TECH_RESEARCH_LABORATORY Technology = 155
	TECH_ROBOMINERS Technology = 156
	TECH_SPACE_SCANNER Technology = 157
	TECH_SCOUT_LAB Technology = 158
	TECH_SECURITY_STATIONS Technology = 159
	TECH_SENSORS Technology = 160
	TECH_SHIELD_CAPACITORS Technology = 161
	TECH_SOIL_ENRICHMENT Technology = 162
	TECH_SPACE_ACADEMY Technology = 163
	TECH_SPACEPORT Technology = 164
	TECH_SPATIAL_COMPRESSOR Technology = 165
	TECH_SPY_NETWORK Technology = 166
	TECH_STANDARD_FUEL_CELLS Technology = 167
	TECH_STAR_BASE Technology = 168
	TECH_STAR_FORTRESS Technology = 169
	TECH_STAR_GATE Technology = 170
	TECH_STASIS_FIELD Technology = 171
	TECH_STEALTH_FIELD Technology = 172
	TECH_STEALTH_SUIT Technology = 173
	TECH_STELLAR_CONVERTER Technology = 174
	TECH_STRUCTURAL_ANALYZER Technology = 175
	TECH_SUBSPACE_COMMUNICATIONS Technology = 176
	TECH_SUBSPACE_TELEPORTER Technology = 177
	TECH_SUBTERRANEAN_FARMS Technology = 178
	TECH_SURVIVAL_PODS Technology = 179
	TECH_TACHYON_COMMUNICATIONS Technology = 180
	TECH_TACHYON_SCANNER Technology = 181
	TECH_TELEPATHIC_TRAINING Technology = 182
	TECH_TERRAFORMING Technology = 183
	TECH_THORIUM_FUEL_CELLS Technology = 184
	TECH_TIME_WARP_FACILITATOR Technology = 185
	TECH_TITAN_CONSTRUCTION Technology = 186
	TECH_TITANIUM_ARMOR Technology = 187
	TECH_TRACTOR_BEAM Technology = 188
	TECH_TRANSPORT Technology = 189
	TECH_TRANSPORTERS Technology = 190
	TECH_TRITANIUM_ARMOR Technology = 191
	TECH_TROOP_PODS Technology = 192
	TECH_UNIVERSAL_ANTIDOTE Technology = 193
	TECH_URRIDIUM_FUEL_CELLS Technology = 194
	TECH_VIRTUAL_REALITY_NETWORK Technology = 195
	TECH_WARP_DISSIPATER Technology = 196
	TECH_WARP_INTERDICTOR Technology = 197
	TECH_WEATHER_CONTROL_SYSTEM Technology = 198
	TECH_WIDE_AREA_JAMMER Technology = 199
	TECH_XENO_PSYCHOLOGY Technology = 200
	TECH_XENTRONIUM_ARMOR Technology = 201
	TECH_ZEON_MISSILE Technology = 202
	TECH_ZORTRIUM_ARMOR Technology = 203
	TECH_HYPER_BIOLOGY Technology = 204
	TECH_HYPER_POWER Technology = 205
	TECH_HYPER_PHYSICS Technology = 206
	TECH_HYPER_CONSTRUCTION Technology = 207
	TECH_HYPER_FIELDS Technology = 208
	TECH_HYPER_CHEMISTRY Technology = 209
	TECH_HYPER_COMPUTERS Technology = 210
	TECH_HYPER_SOCIOLOGY Technology = 211
)

// ColonyBuilding
type ColonyBuilding int
const (
	BUILDING_NONE ColonyBuilding = 0
	BUILDING_ALIEN_CONTROL_CENTER ColonyBuilding = 1
	BUILDING_ARMOR_BARRACKS ColonyBuilding = 2
	BUILDING_ARTEMIS_SYSTEM_NET ColonyBuilding = 3
	BUILDING_ASTRO_UNIVERSITY ColonyBuilding = 4
	BUILDING_ATMOSPHERE_RENEWER ColonyBuilding = 5
	BUILDING_AUTOLAB ColonyBuilding = 6
	BUILDING_AUTOMATED_FACTORY ColonyBuilding = 7
	BUILDING_BATTLESTATION ColonyBuilding = 8
	BUILDING_CAPITOL ColonyBuilding = 9
	BUILDING_CLONING_CENTER ColonyBuilding = 10
	BUILDING_COLONY_BASE ColonyBuilding = 11
	BUILDING_DEEP_CORE_MINE ColonyBuilding = 12
	BUILDING_CORE_WASTE_DUMP ColonyBuilding = 13
	BUILDING_DIMENSIONAL_PORTAL ColonyBuilding = 14
	BUILDING_BIOSPHERES ColonyBuilding = 15
	BUILDING_FOOD_REPLICATORS ColonyBuilding = 16
	BUILDING_GAIA_TRANSFORMATION ColonyBuilding = 17
	BUILDING_CURRENCY_EXCHANGE ColonyBuilding = 18
	BUILDING_GALACTIC_CYBERNET ColonyBuilding = 19
	BUILDING_HOLO_SIMULATOR ColonyBuilding = 20
	BUILDING_HYDROPONIC_FARM ColonyBuilding = 21
	BUILDING_MARINE_BARRACKS ColonyBuilding = 22
	BUILDING_BARRIER_SHIELD ColonyBuilding = 23
	BUILDING_FLUX_SHIELD ColonyBuilding = 24
	BUILDING_GRAVITY_GENERATOR ColonyBuilding = 25
	BUILDING_MISSILE_BASE ColonyBuilding = 26
	BUILDING_GROUND_BATTERIES ColonyBuilding = 27
	BUILDING_RADIATION_SHIELD ColonyBuilding = 28
	BUILDING_STOCK_EXCHANGE ColonyBuilding = 29
	BUILDING_SUPERCOMPUTER ColonyBuilding = 30
	BUILDING_PLEASURE_DOME ColonyBuilding = 31
	BUILDING_POLLUTION_PROCESSOR ColonyBuilding = 32
	BUILDING_RECYCLOTRON ColonyBuilding = 33
	BUILDING_ROBOTIC_FACTORY ColonyBuilding = 34
	BUILDING_RESEARCH_LAB ColonyBuilding = 35
	BUILDING_ROBO_MINER_PLANT ColonyBuilding = 36
	BUILDING_SOIL_ENRICHMENT ColonyBuilding = 37
	BUILDING_SPACE_ACADEMY ColonyBuilding = 38
	BUILDING_SPACEPORT ColonyBuilding = 39
	BUILDING_STAR_BASE ColonyBuilding = 40
	BUILDING_STAR_FORTRESS ColonyBuilding = 41
	BUILDING_STELLAR_CONVERTER ColonyBuilding = 42
	BUILDING_SUBTERRANEAN_FARMS ColonyBuilding = 43
	BUILDING_TERRAFORMING ColonyBuilding = 44
	BUILDING_WARP_INTERDICTOR ColonyBuilding = 45
	BUILDING_WEATHER_CONTROLLER ColonyBuilding = 46
	BUILDING_FIGHTER_GARRISON ColonyBuilding = 47
	BUILDING_ARTIFICIAL_PLANET ColonyBuilding = 48
)

// ForeignPolicy
type ForeignPolicy int
const (
	DIPLO_NONE ForeignPolicy = 0
	DIPLO_NON_AGGRESSION ForeignPolicy = 1
	DIPLO_ALLIANCE ForeignPolicy = 2
	DIPLO_PEACE ForeignPolicy = 3
	DIPLO_LIMITED_WAR ForeignPolicy = 4
	DIPLO_WAR ForeignPolicy = 5
)

// LeaderSkills
type LeaderSkills int
const (
	SKILL_ASSASSIN LeaderSkills = 0
	SKILL_COMMANDO LeaderSkills = 1
	SKILL_DIPLOMAT LeaderSkills = 2
	SKILL_FAMOUS LeaderSkills = 3
	SKILL_MEGAWEALTH LeaderSkills = 4
	SKILL_OPERATIONS LeaderSkills = 5
	SKILL_RESEARCHER LeaderSkills = 6
	SKILL_SPYMASTER LeaderSkills = 7
	SKILL_TELEPATH LeaderSkills = 8
	SKILL_TRADER LeaderSkills = 9
	SKILL_ENGINEER LeaderSkills = 16
	SKILL_FIGHTER_PILOT LeaderSkills = 17
	SKILL_GALACTIC_LORE LeaderSkills = 18
	SKILL_HELMSMAN LeaderSkills = 19
	SKILL_NAVIGATOR LeaderSkills = 20
	SKILL_ORDNANCE LeaderSkills = 21
	SKILL_SECURITY LeaderSkills = 22
	SKILL_WEAPONRY LeaderSkills = 23
	SKILL_ENVIRONMENTALIST LeaderSkills = 32
	SKILL_FARMING_LEADER LeaderSkills = 33
	SKILL_FINANCIAL_LEADER LeaderSkills = 34
	SKILL_INSTRUCTOR LeaderSkills = 35
	SKILL_LABOR_LEADER LeaderSkills = 36
	SKILL_MEDICINE LeaderSkills = 37
	SKILL_SCIENCE_LEADER LeaderSkills = 38
	SKILL_SPIRITUAL_LEADER LeaderSkills = 39
	SKILL_TACTICS LeaderSkills = 40
)

// LeaderState
type LeaderState int
const (
	Dead LeaderState = -2
	Unavailable LeaderState = -1
	Idle LeaderState = 0
	Working LeaderState = 1
	Unassigned LeaderState = 2
	ForHire LeaderState = 4
)

// ShipState
type ShipState int
const (
	InOrbit ShipState = 0
	InTransit ShipState = 1
	LeavingOrbit ShipState = 2
	Unknown ShipState = 3
	InRefit ShipState = 4
	Destroyed ShipState = 5
	UnderConstruction ShipState = 6
)

// ShipType
type ShipType int
const (
	COMBAT_SHIP ShipType = 0
	COLONY_SHIP ShipType = 1
	TRANSPORT_SHIP ShipType = 2
	BAD_SHIP_TYPE ShipType = 3
	OUTPOST_SHIP ShipType = 4
)

// CombatShipClass
type CombatShipClass int
const (
	SHIP_FRIGATE CombatShipClass = 0
	SHIP_DESTROYER CombatShipClass = 1
	SHIP_CRUISER CombatShipClass = 2
	SHIP_BATTLESHIP CombatShipClass = 3
	SHIP_TITAN CombatShipClass = 4
	SHIP_DOOMSTAR CombatShipClass = 5
)

// WeaponArc
type WeaponArc int
const (
	ARC_FWD WeaponArc = 1
	ARC_FWD_EXT WeaponArc = 2
	ARC_BACK_EXT WeaponArc = 4
	ARC_BACK WeaponArc = 8
	ARC_MONSTER_360 WeaponArc = 15
	ARC_360 WeaponArc = 16
)

// SpecialDevices
type SpecialDevices int
const (
	SPEC_ACHILLES_TARGETING_UNIT SpecialDevices = 1
	SPEC_AUGMENTED_ENGINES SpecialDevices = 2
	SPEC_AUTOMATED_REPAIR_UNIT SpecialDevices = 3
	SPEC_BATTLE_PODS SpecialDevices = 4
	SPEC_BATTLE_SCANNER SpecialDevices = 5
	SPEC_CLOAKING_DEVICE SpecialDevices = 6
	SPEC_DAMPER_FIELD SpecialDevices = 7
	SPEC_DISPLACEMENT_DEVICE SpecialDevices = 8
	SPEC_ECM_JAMMER SpecialDevices = 9
	SPEC_ENERGY_ABSORBER SpecialDevices = 10
	SPEC_EXTENDED_FUEL_TANKS SpecialDevices = 11
	SPEC_FAST_MISSILE_RACKS SpecialDevices = 12
	SPEC_HARD_SHIELDS SpecialDevices = 13
	SPEC_HEAVY_ARMOR SpecialDevices = 14
	SPEC_HIGH_ENERGY_FOCUS SpecialDevices = 15
	SPEC_HYPERX_CAPACITORS SpecialDevices = 16
	SPEC_INERTIAL_NULLIFIER SpecialDevices = 17
	SPEC_INERTIAL_STABILIZER SpecialDevices = 18
	SPEC_LIGHTNING_FIELD SpecialDevices = 19
	SPEC_MULTIPHASED_SHIELDS SpecialDevices = 20
	SPEC_MULTIWAVE_ECM_JAMMER SpecialDevices = 21
	SPEC_PHASE_SHIFTER SpecialDevices = 22
	SPEC_PHASING_CLOAK SpecialDevices = 23
	SPEC_QUANTUM_DETONATOR SpecialDevices = 24
	SPEC_RANGEMASTER_UNIT SpecialDevices = 25
	SPEC_REFLECTION_FIELD SpecialDevices = 26
	SPEC_REINFORCED_HULL SpecialDevices = 27
	SPEC_SCOUT_LAB SpecialDevices = 28
	SPEC_SECURITY_STATIONS SpecialDevices = 29
	SPEC_SHIELD_CAPACITOR SpecialDevices = 30
	SPEC_STEALTH_FIELD SpecialDevices = 31
	SPEC_STRUCTURAL_ANALYZER SpecialDevices = 32
	SPEC_SUB_SPACE_TELEPORTER SpecialDevices = 33
	SPEC_TIME_WARP_FACILITATOR SpecialDevices = 34
	SPEC_TRANSPORTERS SpecialDevices = 35
	SPEC_TROOP_PODS SpecialDevices = 36
	SPEC_WARP_DISSIPATOR SpecialDevices = 37
	SPEC_WIDE_AREA_JAMMER SpecialDevices = 38
	SPEC_REGENERATION SpecialDevices = 39
)

// RaceTrait
type RaceTrait int
const (
	TRAIT_GOVERNMENT RaceTrait = 0
	TRAIT_POPULATION RaceTrait = 1
	TRAIT_FARMING RaceTrait = 2
	TRAIT_INDUSTRY RaceTrait = 3
	TRAIT_SCIENCE RaceTrait = 4
	TRAIT_MONEY RaceTrait = 5
	TRAIT_SHIP_DEFENSE RaceTrait = 6
	TRAIT_SHIP_ATTACK RaceTrait = 7
	TRAIT_GROUND_COMBAT RaceTrait = 8
	TRAIT_SPYING RaceTrait = 9
	TRAIT_LOW_G RaceTrait = 10
	TRAIT_HIGH_G RaceTrait = 11
	TRAIT_AQUATIC RaceTrait = 12
	TRAIT_SUBTERRANEAN RaceTrait = 13
	TRAIT_LARGE_HOMEWORLD RaceTrait = 14
	TRAIT_RICH_HOMEWORLD RaceTrait = 15
	TRAIT_ARTIFACTS_HOMEWORLD RaceTrait = 16
	TRAIT_CYBERNETIC RaceTrait = 17
	TRAIT_LITHOVORE RaceTrait = 18
	TRAIT_REPULSIVE RaceTrait = 19
	TRAIT_CHARISMATIC RaceTrait = 20
	TRAIT_UNCREATIVE RaceTrait = 21
	TRAIT_CREATIVE RaceTrait = 22
	TRAIT_TOLERANT RaceTrait = 23
	TRAIT_FANTASTIC_TRADERS RaceTrait = 24
	TRAIT_TELEPATHIC RaceTrait = 25
	TRAIT_LUCKY RaceTrait = 26
	TRAIT_OMNISCIENCE RaceTrait = 27
	TRAIT_STEALTHY_SHIPS RaceTrait = 28
	TRAIT_TRANS_DIMENSIONAL RaceTrait = 29
	TRAIT_WARLORD RaceTrait = 30
	TRAIT_POOR_HOMEWORLD RaceTrait = 31
)

// SelectionFilter
type SelectionFilter int
const (
	SELFILTER_NONE SelectionFilter = 0
	SELFILTER_OWNED SelectionFilter = 1
	SELFILTER_ANY SelectionFilter = 2
)
