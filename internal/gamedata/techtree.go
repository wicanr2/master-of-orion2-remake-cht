// Package gamedata 提供 MOO2 遊戲規則的靜態資料表。
//
// 本檔逐字轉寫自 openorion2 的科技樹資料:
//   - MAX_RESEARCH_TOPICS(83)/MAX_RESEARCH_AREAS(8)/MAX_AREA_TOPICS(14)/MAX_RESEARCH_CHOICES(4):
//     openorion2/src/gamestate.h:62-65, openorion2/src/tech.cpp:40, openorion2/src/tech.h:27
//   - techtree[][]:         openorion2/src/tech.cpp:69-167 (各研究領域含哪些 ResearchTopic)
//   - research_choices[83]: openorion2/src/tech.cpp:169-305 (每個 topic 的花費與可選科技)
//
// Technology / ResearchTopic 兩個 enum 型別與其常數已由 enums.go 從 gamestate.h 產生
// (TECH_* / TOPIC_* / RESEARCH_*,整數值與 C 版完全一致),本檔直接沿用,不重複宣告,
// 避免與既有生成檔重複定義。
//
// 數值為機械轉寫,忠實照抄原始碼常數與陣列內容,未做任何調整或最佳化。
package gamedata

// ResearchChoice 對應 C 版 struct ResearchChoice (tech.h:29-33)。
// ResearchAll 對應原本的 int research_all (0/1 轉 bool)。
// Choices 只放實際可選科技,原始 C 陣列中用來補滿 MAX_RESEARCH_CHOICES 的 TECH_NONE 填充值不放進來。
type ResearchChoice struct {
	Cost        int
	ResearchAll bool
	Choices     []Technology
}

// researchChoices 逐字轉寫自 tech.cpp:169-305 的 research_choices[MAX_RESEARCH_TOPICS] (83 列)。
// 陣列索引即為 ResearchTopic 的整數值 (與 C 版 research_choices[topic] 存取方式一致)。
var researchChoices = [83]ResearchChoice{
	{Cost: 0, ResearchAll: false}, // TOPIC_STARTING_TECH
	{Cost: 400, ResearchAll: false, Choices: []Technology{TECH_CLONING_CENTER, TECH_DEATH_SPORES, TECH_SOIL_ENRICHMENT}},                                // TOPIC_ADVANCED_BIOLOGY
	{Cost: 650, ResearchAll: false, Choices: []Technology{TECH_MERCULITE_MISSILE, TECH_POLLUTION_PROCESSOR}},                                            // TOPIC_ADVANCED_CHEMISTRY
	{Cost: 150, ResearchAll: false, Choices: []Technology{TECH_AUTOMATED_FACTORIES, TECH_HEAVY_ARMOR, TECH_PLANETARY_MISSILE_BASE}},                     // TOPIC_ADVANCED_CONSTRUCTION
	{Cost: 80, ResearchAll: false, Choices: []Technology{TECH_ANTIMISSILE_ROCKETS, TECH_REINFORCED_HULL, TECH_FIGHTER_BAYS}},                            // TOPIC_ADVANCED_ENGINEERING
	{Cost: 250, ResearchAll: false, Choices: []Technology{TECH_AUGMENTED_ENGINES, TECH_FUSION_BOMB, TECH_FUSION_DRIVE}},                                 // TOPIC_ADVANCED_FUSION
	{Cost: 4500, ResearchAll: false, Choices: []Technology{TECH_CONFEDERATION, TECH_FEDERATION, TECH_GALACTIC_UNIFICATION, TECH_IMPERIUM}},              // TOPIC_ADVANCED_GOVERNMENTS
	{Cost: 250, ResearchAll: false, Choices: []Technology{TECH_CLASS_I_SHIELD, TECH_ECM_JAMMER, TECH_MASS_DRIVER}},                                      // TOPIC_ADVANCED_MAGNETISM
	{Cost: 1500, ResearchAll: false, Choices: []Technology{TECH_PLANET_CONSTRUCTION, TECH_AUTOMATED_REPAIR_UNIT, TECH_RECYCLOTRON}},                     // TOPIC_ADVANCED_MANUFACTURING
	{Cost: 250, ResearchAll: false, Choices: []Technology{TECH_DEUTERIUM_FUEL_CELLS, TECH_TRITANIUM_ARMOR}},                                             // TOPIC_ADVANCED_METALLURGY
	{Cost: 150, ResearchAll: false, Choices: []Technology{TECH_SPACE_ACADEMY}},                                                                          // TOPIC_MILITARY_TACTICS
	{Cost: 2000, ResearchAll: false, Choices: []Technology{TECH_BOMBER_BAYS, TECH_ROBOTIC_FACTORY}},                                                     // TOPIC_ADVANCED_ROBOTICS
	{Cost: 2000, ResearchAll: false, Choices: []Technology{TECH_ASTRO_UNIVERSITY}},                                                                      // TOPIC_TEACHING_METHODS
	{Cost: 2000, ResearchAll: false, Choices: []Technology{TECH_ANTIMATTER_BOMB, TECH_ANTIMATTER_DRIVE, TECH_ANTIMATTER_TORPEDOES}},                     // TOPIC_ANTIMATTER_FISSION
	{Cost: 1500, ResearchAll: false, Choices: []Technology{TECH_CYBERSECURITY_LINK, TECH_EMISSIONS_GUIDANCE_SYSTEM, TECH_RANGEMASTER_UNIT}},             // TOPIC_ARTIFICIAL_CONSCIOUSNESS
	{Cost: 400, ResearchAll: false, Choices: []Technology{TECH_NEURAL_SCANNER, TECH_SCOUT_LAB, TECH_SECURITY_STATIONS}},                                 // TOPIC_ARTIFICIAL_INTELLIGENCE
	{Cost: 1150, ResearchAll: false, Choices: []Technology{TECH_GRAVITON_BEAM, TECH_PLANETARY_GRAVITY_GENERATOR, TECH_TRACTOR_BEAM}},                    // TOPIC_ARTIFICIAL_GRAVITY
	{Cost: 4500, ResearchAll: false, Choices: []Technology{TECH_BIOTERMINATOR, TECH_UNIVERSAL_ANTIDOTE}},                                                // TOPIC_ARTIFICIAL_LIFE
	{Cost: 80, ResearchAll: false, Choices: []Technology{TECH_BIOSPHERES, TECH_HYDROPONIC_FARM}},                                                        // TOPIC_ASTRO_BIOLOGY
	{Cost: 1150, ResearchAll: false, Choices: []Technology{TECH_BATTLEOIDS, TECH_GROUND_BATTERIES, TECH_TITAN_CONSTRUCTION}},                            // TOPIC_ASTRO_CONSTRUCTION
	{Cost: 400, ResearchAll: false, Choices: []Technology{TECH_ARMOR_BARRACKS, TECH_FIGHTER_GARRISON, TECH_SPACEPORT}},                                  // TOPIC_ASTRO_ENGINEERING
	{Cost: 250, ResearchAll: false, Choices: []Technology{TECH_BATTLE_PODS, TECH_SURVIVAL_PODS, TECH_TROOP_PODS}},                                       // TOPIC_CAPSULE_CONSTRUCTION
	{Cost: 50, ResearchAll: true, Choices: []Technology{TECH_EXTENDED_FUEL_TANKS, TECH_NUCLEAR_MISSILE, TECH_STANDARD_FUEL_CELLS, TECH_TITANIUM_ARMOR}}, // TOPIC_CHEMISTRY
	{Cost: 80, ResearchAll: false, Choices: []Technology{TECH_COLONY_SHIP, TECH_OUTPOST_SHIP, TECH_TRANSPORT}},                                          // TOPIC_COLD_FUSION
	{Cost: 3500, ResearchAll: false, Choices: []Technology{TECH_ANDROID_FARMERS, TECH_ANDROID_SCIENTISTS, TECH_ANDROID_WORKERS}},                        // TOPIC_CYBERTECHNICS
	{Cost: 2750, ResearchAll: false, Choices: []Technology{TECH_AUTOLAB, TECH_CYBERTRONIC_COMPUTER, TECH_STRUCTURAL_ANALYZER}},                          // TOPIC_CYBERTRONICS
	{Cost: 3500, ResearchAll: false, Choices: []Technology{TECH_CLOAKING_DEVICE, TECH_HARD_SHIELDS, TECH_STASIS_FIELD}},                                 // TOPIC_DISTORTION_FIELDS
	{Cost: 1500, ResearchAll: false, Choices: []Technology{TECH_PERSONAL_SHIELD, TECH_STEALTH_FIELD, TECH_STEALTH_SUIT}},                                // TOPIC_ELECTROMAGNETIC_REFRACTION
	{Cost: 50, ResearchAll: true, Choices: []Technology{TECH_ELECTRONIC_COMPUTER}},                                                                      // TOPIC_ELECTRONICS
	{Cost: 50, ResearchAll: true, Choices: []Technology{TECH_COLONY_BASE, TECH_MARINE_BARRACKS, TECH_STAR_BASE}},                                        // TOPIC_ENGINEERING
	{Cost: 2750, ResearchAll: false, Choices: []Technology{TECH_HEIGHTENED_INTELLIGENCE, TECH_PSIONICS}},                                                // TOPIC_EVOLUTIONARY_GENETICS
	{Cost: 150, ResearchAll: false, Choices: []Technology{TECH_FUSION_BEAM, TECH_FUSION_RIFLE}},                                                         // TOPIC_FUSION_PHYSICS
	{Cost: 6000, ResearchAll: false, Choices: []Technology{TECH_GALACTIC_CURRENCY_EXCHANGE}},                                                            // TOPIC_GALACTIC_ECONOMICS
	{Cost: 4500, ResearchAll: false, Choices: []Technology{TECH_GALACTIC_CYBERNET, TECH_VIRTUAL_REALITY_NETWORK}},                                       // TOPIC_GALACTIC_NETWORKING
	{Cost: 900, ResearchAll: false, Choices: []Technology{TECH_MICROBIOTICS, TECH_TELEPATHIC_TRAINING}},                                                 // TOPIC_GENETIC_ENGINEERING
	{Cost: 1150, ResearchAll: false, Choices: []Technology{TECH_TERRAFORMING}},                                                                          // TOPIC_GENETIC_MUTATIONS
	{Cost: 650, ResearchAll: false, Choices: []Technology{TECH_ANTIGRAV_HARNESS, TECH_GYRO_DESTABILIZER, TECH_INERTIAL_STABILIZER}},                     // TOPIC_GRAVITIC_FIELDS
	{Cost: 3500, ResearchAll: false, Choices: []Technology{TECH_ENERGY_ABSORBER, TECH_HIGH_ENERGY_FOCUS, TECH_MEGAFLUXERS}},                             // TOPIC_HIGH_ENERGY_DISTRIBUTION
	{Cost: 4500, ResearchAll: false, Choices: []Technology{TECH_HYPER_DRIVE, TECH_HYPERX_CAPACITORS, TECH_PROTON_TORPEDOES}},                            // TOPIC_HYPER_DIMENSIONAL_FISSION
	{Cost: 6000, ResearchAll: false, Choices: []Technology{TECH_HYPERSPACE_COMMUNICATIONS, TECH_MAULER_DEVICE, TECH_SENSORS}},                           // TOPIC_HYPER_DIMENSIONAL_PHYSICS
	{Cost: 10000, ResearchAll: false, Choices: []Technology{TECH_INTERPHASED_DRIVE, TECH_NEUTRONIUM_BOMB, TECH_PLASMA_TORPEDOES}},                       // TOPIC_INTERPHASED_FISSION
	{Cost: 900, ResearchAll: false, Choices: []Technology{TECH_ION_DRIVE, TECH_ION_PULSE_CANNON, TECH_SHIELD_CAPACITORS}},                               // TOPIC_ION_FISSION
	{Cost: 6000, ResearchAll: false, Choices: []Technology{TECH_ADVANCED_CITY_PLANNING, TECH_HEAVY_FIGHTER_BAYS, TECH_STAR_FORTRESS}},                   // TOPIC_SUPERSCALAR_CONSTRUCTION
	{Cost: 1150, ResearchAll: false, Choices: []Technology{TECH_PLANETARY_STOCK_EXCHANGE}},                                                              // TOPIC_MACRO_ECONOMICS
	{Cost: 1500, ResearchAll: false, Choices: []Technology{TECH_SUBTERRANEAN_FARMS, TECH_WEATHER_CONTROL_SYSTEM}},                                       // TOPIC_MACRO_GENETICS
	{Cost: 900, ResearchAll: false, Choices: []Technology{TECH_CLASS_III_SHIELD, TECH_PLANETARY_RADIATION_SHIELD, TECH_WARP_DISSIPATER}},                // TOPIC_MAGNETO_GRAVITICS
	{Cost: 2750, ResearchAll: false, Choices: []Technology{TECH_FOOD_REPLICATORS, TECH_TRANSPORTERS}},                                                   // TOPIC_MATTER_ENERGY_CONVERSION
	{Cost: 1150, ResearchAll: false, Choices: []Technology{TECH_ATMOSPHERIC_RENEWER, TECH_IRIDIUM_FUEL_CELLS, TECH_PULSON_MISSILE}},                     // TOPIC_MOLECULAR_COMPRESSION
	{Cost: 10000, ResearchAll: false, Choices: []Technology{TECH_ADAMANTIUM_ARMOR, TECH_THORIUM_FUEL_CELLS}},                                            // TOPIC_MOLECULAR_CONTROL
	{Cost: 6000, ResearchAll: false, Choices: []Technology{TECH_ACHILLES_TARGETING_UNIT, TECH_MOLECULARTRONIC_COMPUTER, TECH_PLEASURE_DOME}},            // TOPIC_MOLECULATRONICS
	{Cost: 4500, ResearchAll: false, Choices: []Technology{TECH_NEUTRONIUM_ARMOR, TECH_URRIDIUM_FUEL_CELLS, TECH_ZEON_MISSILE}},                         // TOPIC_MOLECULAR_MANIPULATION
	{Cost: 4500, ResearchAll: false, Choices: []Technology{TECH_DIMENSIONAL_PORTAL, TECH_DISRUPTER_CANNON}},                                             // TOPIC_MULTIDIMENSIONAL_PHYSICS
	{Cost: 2000, ResearchAll: false, Choices: []Technology{TECH_MULTIPHASED_SHIELDS, TECH_PHASOR, TECH_PHASOR_RIFLE}},                                   // TOPIC_MULTIPHASED_PHYSICS
	{Cost: 2000, ResearchAll: false, Choices: []Technology{TECH_MICROLITE_CONSTRUCTION, TECH_NANO_DISASSEMBLERS, TECH_ZORTRIUM_ARMOR}},                  // TOPIC_NANO_TECHNOLOGY
	{Cost: 900, ResearchAll: false, Choices: []Technology{TECH_NEUTRON_BLASTER, TECH_NEUTRON_SCANNER}},                                                  // TOPIC_NEUTRINO_PHYSICS
	{Cost: 50, ResearchAll: true, Choices: []Technology{TECH_FREIGHTERS, TECH_NUCLEAR_BOMB, TECH_NUCLEAR_DRIVE}},                                        // TOPIC_NUCLEAR_FISSION
	{Cost: 150, ResearchAll: false, Choices: []Technology{TECH_DAUNTLESS_GUIDANCE_SYSTEM, TECH_OPTRONIC_COMPUTER, TECH_RESEARCH_LABORATORY}},            // TOPIC_OPTRONICS
	{Cost: 50, ResearchAll: true, Choices: []Technology{TECH_LASER_CANNON, TECH_LASER_RIFLE, TECH_SPACE_SCANNER}},                                       // TOPIC_PHYSICS
	{Cost: 7500, ResearchAll: false, Choices: []Technology{TECH_ARTEMIS_SYSTEM_NET, TECH_DOOM_STAR_CONSTRUCTION}},                                       // TOPIC_PLANETOID_CONSTRUCTION
	{Cost: 3500, ResearchAll: false, Choices: []Technology{TECH_PLASMA_CANNON, TECH_PLASMA_RIFLE, TECH_PLASMA_WEB}},                                     // TOPIC_PLASMA_PHYSICS
	{Cost: 900, ResearchAll: false, Choices: []Technology{TECH_HOLO_SIMULATOR, TECH_PLANETARY_SUPERCOMPUTER, TECH_POSITRONIC_COMPUTER}},                 // TOPIC_POSITRONICS
	{Cost: 4500, ResearchAll: false, Choices: []Technology{TECH_CLASS_VII_SHIELD, TECH_PLANETARY_FLUX_SHIELD, TECH_WIDE_AREA_JAMMER}},                   // TOPIC_QUANTUM_FIELDS
	{Cost: 650, ResearchAll: false, Choices: []Technology{TECH_BATTLESTATION, TECH_POWERED_ARMOR, TECH_ROBOMINERS}},                                     // TOPIC_ROBOTICS
	{Cost: 900, ResearchAll: false, Choices: []Technology{TECH_ADVANCED_DAMAGE_CONTROL, TECH_ASSAULT_SHUTTLES, TECH_FAST_MISSILE_RACKS}},                // TOPIC_SERVO_MECHANICS
	{Cost: 2750, ResearchAll: false, Choices: []Technology{TECH_CLASS_V_SHIELD, TECH_GAUSS_CANNON, TECH_MULTIWAVE_ECM_JAMMER}},                          // TOPIC_SUBSPACE_FIELDS
	{Cost: 1500, ResearchAll: false, Choices: []Technology{TECH_JUMP_GATE, TECH_SUBSPACE_COMMUNICATIONS}},                                               // TOPIC_SUBSPACE_PHYSICS
	{Cost: 250, ResearchAll: false, Choices: []Technology{TECH_BATTLE_SCANNER, TECH_TACHYON_COMMUNICATIONS, TECH_TACHYON_SCANNER}},                      // TOPIC_TACHYON_PHYSICS
	{Cost: 3500, ResearchAll: false, Choices: []Technology{TECH_DEEP_CORE_MINING, TECH_CORE_WASTE_DUMPS}},                                               // TOPIC_TECTONIC_ENGINEERING
	{Cost: 15000, ResearchAll: false, Choices: []Technology{TECH_CLASS_X_SHIELD, TECH_PHASING_CLOAK, TECH_PLANETARY_BARRIER_SHIELD}},                    // TOPIC_TEMPORAL_FIELDS
	{Cost: 15000, ResearchAll: false, Choices: []Technology{TECH_STAR_GATE, TECH_STELLAR_CONVERTER, TECH_TIME_WARP_FACILITATOR}},                        // TOPIC_TEMPORAL_PHYSICS
	{Cost: 7500, ResearchAll: false, Choices: []Technology{TECH_BIOMORPHIC_FUNGI, TECH_EVOLUTIONARY_MUTATION, TECH_GAIA_TRANSFORMATION}},                // TOPIC_TRANS_GENETICS
	{Cost: 7500, ResearchAll: false, Choices: []Technology{TECH_DISPLACEMENT_DEVICE, TECH_INERTIAL_NULLIFIER, TECH_SUBSPACE_TELEPORTER}},                // TOPIC_TRANSWARP_FIELDS
	{Cost: 2000, ResearchAll: false, Choices: []Technology{TECH_LIGHTNING_FIELD, TECH_PULSAR, TECH_WARP_INTERDICTOR}},                                   // TOPIC_WARP_FIELDS
	{Cost: 650, ResearchAll: false, Choices: []Technology{TECH_ALIEN_MANAGEMENT_CENTER, TECH_XENO_PSYCHOLOGY}},                                          // TOPIC_XENO_RELATIONS
	{Cost: 0, ResearchAll: false}, // TOPIC_XENON_TECHNOLOGY
	{Cost: 25000, ResearchAll: false, Choices: []Technology{TECH_HYPER_BIOLOGY}},      // TOPIC_HYPER_BIOLOGY
	{Cost: 25000, ResearchAll: false, Choices: []Technology{TECH_HYPER_POWER}},        // TOPIC_HYPER_POWER
	{Cost: 25000, ResearchAll: false, Choices: []Technology{TECH_HYPER_PHYSICS}},      // TOPIC_HYPER_PHYSICS
	{Cost: 25000, ResearchAll: false, Choices: []Technology{TECH_HYPER_CONSTRUCTION}}, // TOPIC_HYPER_CONSTRUCTION
	{Cost: 25000, ResearchAll: false, Choices: []Technology{TECH_HYPER_FIELDS}},       // TOPIC_HYPER_FIELDS
	{Cost: 25000, ResearchAll: false, Choices: []Technology{TECH_HYPER_CHEMISTRY}},    // TOPIC_HYPER_CHEMISTRY
	{Cost: 25000, ResearchAll: false, Choices: []Technology{TECH_HYPER_COMPUTERS}},    // TOPIC_HYPER_COMPUTERS
	{Cost: 25000, ResearchAll: false, Choices: []Technology{TECH_HYPER_SOCIOLOGY}},    // TOPIC_HYPER_SOCIOLOGY
}

// ResearchChoiceFor 回傳指定 ResearchTopic 的花費與可選科技清單。
func ResearchChoiceFor(topic ResearchTopic) ResearchChoice {
	return researchChoices[topic]
}

// techtree 逐字轉寫自 tech.cpp:69-167 的 techtree[MAX_RESEARCH_AREAS][MAX_AREA_TOPICS]。
// 索引順序與 enums.go 的 ResearchArea 常數值一致 (RESEARCH_BIOLOGY=0 ... RESEARCH_SOCIOLOGY=7);
// 每個研究領域含哪些 ResearchTopic,依 C 陣列原始順序照抄;C 陣列中用來補滿 MAX_AREA_TOPICS 的
// TOPIC_STARTING_TECH(0) 填充值不放進來。
var techtree = [8][]ResearchTopic{
	{TOPIC_ASTRO_BIOLOGY, TOPIC_ADVANCED_BIOLOGY, TOPIC_GENETIC_ENGINEERING, TOPIC_GENETIC_MUTATIONS, TOPIC_MACRO_GENETICS, TOPIC_EVOLUTIONARY_GENETICS, TOPIC_ARTIFICIAL_LIFE, TOPIC_TRANS_GENETICS, TOPIC_HYPER_BIOLOGY},                                                                                                                                                        // Biology (RESEARCH_BIOLOGY)
	{TOPIC_NUCLEAR_FISSION, TOPIC_COLD_FUSION, TOPIC_ADVANCED_FUSION, TOPIC_ION_FISSION, TOPIC_ANTIMATTER_FISSION, TOPIC_MATTER_ENERGY_CONVERSION, TOPIC_HIGH_ENERGY_DISTRIBUTION, TOPIC_HYPER_DIMENSIONAL_FISSION, TOPIC_INTERPHASED_FISSION, TOPIC_HYPER_POWER},                                                                                                                 // Power (RESEARCH_POWER)
	{TOPIC_PHYSICS, TOPIC_FUSION_PHYSICS, TOPIC_TACHYON_PHYSICS, TOPIC_NEUTRINO_PHYSICS, TOPIC_ARTIFICIAL_GRAVITY, TOPIC_SUBSPACE_PHYSICS, TOPIC_MULTIPHASED_PHYSICS, TOPIC_PLASMA_PHYSICS, TOPIC_MULTIDIMENSIONAL_PHYSICS, TOPIC_HYPER_DIMENSIONAL_PHYSICS, TOPIC_TEMPORAL_PHYSICS, TOPIC_HYPER_PHYSICS},                                                                         // Physics (RESEARCH_PHYSICS)
	{TOPIC_ENGINEERING, TOPIC_ADVANCED_ENGINEERING, TOPIC_ADVANCED_CONSTRUCTION, TOPIC_CAPSULE_CONSTRUCTION, TOPIC_ASTRO_ENGINEERING, TOPIC_ROBOTICS, TOPIC_SERVO_MECHANICS, TOPIC_ASTRO_CONSTRUCTION, TOPIC_ADVANCED_MANUFACTURING, TOPIC_ADVANCED_ROBOTICS, TOPIC_TECTONIC_ENGINEERING, TOPIC_SUPERSCALAR_CONSTRUCTION, TOPIC_PLANETOID_CONSTRUCTION, TOPIC_HYPER_CONSTRUCTION}, // Construction (RESEARCH_CONSTRUCTION)
	{TOPIC_ADVANCED_MAGNETISM, TOPIC_GRAVITIC_FIELDS, TOPIC_MAGNETO_GRAVITICS, TOPIC_ELECTROMAGNETIC_REFRACTION, TOPIC_WARP_FIELDS, TOPIC_SUBSPACE_FIELDS, TOPIC_DISTORTION_FIELDS, TOPIC_QUANTUM_FIELDS, TOPIC_TRANSWARP_FIELDS, TOPIC_TEMPORAL_FIELDS, TOPIC_HYPER_FIELDS},                                                                                                      // Force Fields (RESEARCH_FIELDS)
	{TOPIC_CHEMISTRY, TOPIC_ADVANCED_METALLURGY, TOPIC_ADVANCED_CHEMISTRY, TOPIC_MOLECULAR_COMPRESSION, TOPIC_NANO_TECHNOLOGY, TOPIC_MOLECULAR_MANIPULATION, TOPIC_MOLECULAR_CONTROL, TOPIC_HYPER_CHEMISTRY},                                                                                                                                                                      // Chemistry (RESEARCH_CHEMISTRY)
	{TOPIC_ELECTRONICS, TOPIC_OPTRONICS, TOPIC_ARTIFICIAL_INTELLIGENCE, TOPIC_POSITRONICS, TOPIC_ARTIFICIAL_CONSCIOUSNESS, TOPIC_CYBERTRONICS, TOPIC_CYBERTECHNICS, TOPIC_GALACTIC_NETWORKING, TOPIC_MOLECULATRONICS, TOPIC_HYPER_COMPUTERS},                                                                                                                                      // Computers (RESEARCH_COMPUTERS)
	{TOPIC_MILITARY_TACTICS, TOPIC_XENO_RELATIONS, TOPIC_MACRO_ECONOMICS, TOPIC_TEACHING_METHODS, TOPIC_ADVANCED_GOVERNMENTS, TOPIC_GALACTIC_ECONOMICS, TOPIC_HYPER_SOCIOLOGY},                                                                                                                                                                                                    // Sociology (RESEARCH_SOCIOLOGY)
}

// TechTree 回傳各研究領域的 ResearchTopic 清單,索引順序對應 enums.go 的 ResearchArea 常數
// (RESEARCH_BIOLOGY=0 ... RESEARCH_SOCIOLOGY=7)。
func TechTree() [][]ResearchTopic {
	result := make([][]ResearchTopic, len(techtree))
	for i, area := range techtree {
		result[i] = append([]ResearchTopic(nil), area...)
	}
	return result
}
