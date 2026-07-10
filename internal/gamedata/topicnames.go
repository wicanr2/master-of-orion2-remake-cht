package gamedata

import "fmt"

// topicnames.go:83 個 ResearchTopic 的英文顯示名表。
//
// 來源與建法(2026-07-10):
//   - topic 0..74 的英文名取自 openorion2 techname 顯示字串(TECHNAME.LBX,同 tech.tsv
//     類別「研究主題」的 75 條 key),與 enums.go 的 TOPIC_* 整數值一一對應。
//     配對經腳本以「enum index == tech.tsv 研究主題出現序」逐項斷言核對(75/75 相符)。
//   - 名稱刻意採 TECHNAME.LBX 的實際拼寫(含連字號:Anti-Matter Fission /
//     Multi-Dimensional Physics / Multi-Phased Physics),使其正好等於 assets/i18n/tech.tsv
//     的 key,顯示層 i18n 才查得到中文。**不用程式化 title-case**——那會把
//     "Anti-Matter Fission" 誤產成 "Antimatter Fission" 而對不上 tech.tsv。
//   - topic 75..82(TOPIC_HYPER_*):原版對所有 hyper 主題共用單一字串
//     ESTR_RTOPIC_HYPER("未來科技"佔位),TECHNAME.LBX 無個別顯示名。本表為 remake
//     依 enum 常數名給各自英文名(Hyper Biology…Hyper Sociology),tech.tsv 同步補譯。
//
// 這是「topic → 英文顯示名(= i18n key)」的權威來源;shell 層不 import i18n,
// 由 cmd/moo2 顯示端經 catalog 翻中文(見 ResearchTopicName / researchchoice.go)。
var topicEnglishNames = map[ResearchTopic]string{
	TOPIC_STARTING_TECH:              "Starting Tech",
	TOPIC_ADVANCED_BIOLOGY:           "Advanced Biology",
	TOPIC_ADVANCED_CHEMISTRY:         "Advanced Chemistry",
	TOPIC_ADVANCED_CONSTRUCTION:      "Advanced Construction",
	TOPIC_ADVANCED_ENGINEERING:       "Advanced Engineering",
	TOPIC_ADVANCED_FUSION:            "Advanced Fusion",
	TOPIC_ADVANCED_GOVERNMENTS:       "Advanced Governments",
	TOPIC_ADVANCED_MAGNETISM:         "Advanced Magnetism",
	TOPIC_ADVANCED_MANUFACTURING:     "Advanced Manufacturing",
	TOPIC_ADVANCED_METALLURGY:        "Advanced Metallurgy",
	TOPIC_MILITARY_TACTICS:           "Military Tactics",
	TOPIC_ADVANCED_ROBOTICS:          "Advanced Robotics",
	TOPIC_TEACHING_METHODS:           "Teaching Methods",
	TOPIC_ANTIMATTER_FISSION:         "Anti-Matter Fission",
	TOPIC_ARTIFICIAL_CONSCIOUSNESS:   "Artificial Consciousness",
	TOPIC_ARTIFICIAL_INTELLIGENCE:    "Artificial Intelligence",
	TOPIC_ARTIFICIAL_GRAVITY:         "Artificial Gravity",
	TOPIC_ARTIFICIAL_LIFE:            "Artificial Life",
	TOPIC_ASTRO_BIOLOGY:              "Astro Biology",
	TOPIC_ASTRO_CONSTRUCTION:         "Astro Construction",
	TOPIC_ASTRO_ENGINEERING:          "Astro Engineering",
	TOPIC_CAPSULE_CONSTRUCTION:       "Capsule Construction",
	TOPIC_CHEMISTRY:                  "Chemistry",
	TOPIC_COLD_FUSION:                "Cold Fusion",
	TOPIC_CYBERTECHNICS:              "Cybertechnics",
	TOPIC_CYBERTRONICS:               "Cybertronics",
	TOPIC_DISTORTION_FIELDS:          "Distortion Fields",
	TOPIC_ELECTROMAGNETIC_REFRACTION: "Electromagnetic Refraction",
	TOPIC_ELECTRONICS:                "Electronics",
	TOPIC_ENGINEERING:                "Engineering",
	TOPIC_EVOLUTIONARY_GENETICS:      "Evolutionary Genetics",
	TOPIC_FUSION_PHYSICS:             "Fusion Physics",
	TOPIC_GALACTIC_ECONOMICS:         "Galactic Economics",
	TOPIC_GALACTIC_NETWORKING:        "Galactic Networking",
	TOPIC_GENETIC_ENGINEERING:        "Genetic Engineering",
	TOPIC_GENETIC_MUTATIONS:          "Genetic Mutations",
	TOPIC_GRAVITIC_FIELDS:            "Gravitic Fields",
	TOPIC_HIGH_ENERGY_DISTRIBUTION:   "High Energy Distribution",
	TOPIC_HYPER_DIMENSIONAL_FISSION:  "Hyper Dimensional Fission",
	TOPIC_HYPER_DIMENSIONAL_PHYSICS:  "Hyper Dimensional Physics",
	TOPIC_INTERPHASED_FISSION:        "Interphased Fission",
	TOPIC_ION_FISSION:                "Ion Fission",
	TOPIC_SUPERSCALAR_CONSTRUCTION:   "Superscalar Construction",
	TOPIC_MACRO_ECONOMICS:            "Macro Economics",
	TOPIC_MACRO_GENETICS:             "Macro Genetics",
	TOPIC_MAGNETO_GRAVITICS:          "Magneto Gravitics",
	TOPIC_MATTER_ENERGY_CONVERSION:   "Matter Energy Conversion",
	TOPIC_MOLECULAR_COMPRESSION:      "Molecular Compression",
	TOPIC_MOLECULAR_CONTROL:          "Molecular Control",
	TOPIC_MOLECULATRONICS:            "Moleculatronics",
	TOPIC_MOLECULAR_MANIPULATION:     "Molecular Manipulation",
	TOPIC_MULTIDIMENSIONAL_PHYSICS:   "Multi-Dimensional Physics",
	TOPIC_MULTIPHASED_PHYSICS:        "Multi-Phased Physics",
	TOPIC_NANO_TECHNOLOGY:            "Nano Technology",
	TOPIC_NEUTRINO_PHYSICS:           "Neutrino Physics",
	TOPIC_NUCLEAR_FISSION:            "Nuclear Fission",
	TOPIC_OPTRONICS:                  "Optronics",
	TOPIC_PHYSICS:                    "Physics",
	TOPIC_PLANETOID_CONSTRUCTION:     "Planetoid Construction",
	TOPIC_PLASMA_PHYSICS:             "Plasma Physics",
	TOPIC_POSITRONICS:                "Positronics",
	TOPIC_QUANTUM_FIELDS:             "Quantum Fields",
	TOPIC_ROBOTICS:                   "Robotics",
	TOPIC_SERVO_MECHANICS:            "Servo Mechanics",
	TOPIC_SUBSPACE_FIELDS:            "Subspace Fields",
	TOPIC_SUBSPACE_PHYSICS:           "Subspace Physics",
	TOPIC_TACHYON_PHYSICS:            "Tachyon Physics",
	TOPIC_TECTONIC_ENGINEERING:       "Tectonic Engineering",
	TOPIC_TEMPORAL_FIELDS:            "Temporal Fields",
	TOPIC_TEMPORAL_PHYSICS:           "Temporal Physics",
	TOPIC_TRANS_GENETICS:             "Trans Genetics",
	TOPIC_TRANSWARP_FIELDS:           "Transwarp Fields",
	TOPIC_WARP_FIELDS:                "Warp Fields",
	TOPIC_XENO_RELATIONS:             "Xeno Relations",
	TOPIC_XENON_TECHNOLOGY:           "Xenon Technology",
	// TOPIC_HYPER_*(75..82):原版共用 ESTR_RTOPIC_HYPER 佔位;remake 給各自英文名。
	TOPIC_HYPER_BIOLOGY:      "Hyper Biology",
	TOPIC_HYPER_POWER:        "Hyper Power",
	TOPIC_HYPER_PHYSICS:      "Hyper Physics",
	TOPIC_HYPER_CONSTRUCTION: "Hyper Construction",
	TOPIC_HYPER_FIELDS:       "Hyper Fields",
	TOPIC_HYPER_CHEMISTRY:    "Hyper Chemistry",
	TOPIC_HYPER_COMPUTERS:    "Hyper Computers",
	TOPIC_HYPER_SOCIOLOGY:    "Hyper Sociology",
}

// TopicEnglishName 回傳研究主題的英文顯示名(= assets/i18n/tech.tsv 的 key)。
// 查無(理論上不會發生,83 個全收錄)回 fmt.Sprintf("Topic#%d", ...)。
func TopicEnglishName(t ResearchTopic) string {
	if name, ok := topicEnglishNames[t]; ok {
		return name
	}
	return fmt.Sprintf("Topic#%d", int(t))
}
