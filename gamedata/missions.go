package gamedata

// MissionID identifies one mission. Explicit literals, append-only, never
// reused: every location id in a seed is derived from one of these.
type MissionID uint16

// Mission is one .pop file: an ordered run of waves on one map, at one
// difficulty tier.
type Mission struct {
	ID         MissionID
	PopFile    string
	Name       string
	Map        MapID
	Difficulty Difficulty
	Waves      uint8

	// HasTank says a tank appears somewhere in the mission, which is what
	// makes the tank check reachable. Three of Valve's missions have none.
	//
	// A false here costs one check. A true on a mission with no tank costs the
	// seed: the location can never be reached and the run cannot be finished.
	// So the source is the wiki's tank health table, which lists every tank
	// spawn by mission and wave, and anything absent from it is false.
	HasTank bool
}

// Missions is the 29 Valve missions. Pop file names come from
// tf2_misc_dir.vpk; display names, tiers and wave counts from the TF2 wiki.
//
// UNVERIFIED: inside a tier with more than one entry, the pairing of display
// name to pop file is a guess. Settle it against resource/tf_english.txt in the
// VPK before the first seed is played: the name is baked into location names.
//
// Community missions are out of scope for v1: nothing on the host can read a
// wave count out of a .pop file, and a downloaded pack can change under a seed
// that already references it.
var Missions = []Mission{
	{1, "mvm_decoy", "Doe's Drill", MapDecoy, DifficultyNormal, 8, true},
	{2, "mvm_decoy_intermediate", "Doe's Doom", MapDecoy, DifficultyIntermediate, 7, true},
	{3, "mvm_decoy_intermediate2", "Day of Wreckening", MapDecoy, DifficultyIntermediate, 6, true},
	{4, "mvm_decoy_advanced", "Disk Deletion", MapDecoy, DifficultyAdvanced, 8, true},
	{5, "mvm_decoy_advanced2", "Data Demolition", MapDecoy, DifficultyAdvanced, 6, true},
	{6, "mvm_decoy_advanced3", "Disintegration", MapDecoy, DifficultyAdvanced, 6, true},
	{7, "mvm_decoy_expert1", "Desperation", MapDecoy, DifficultyExpert, 7, true},
	{8, "mvm_coaltown", "Crash Course", MapCoaltown, DifficultyNormal, 6, true},
	{9, "mvm_coaltown_intermediate", "Cave-in", MapCoaltown, DifficultyIntermediate, 6, true},
	{10, "mvm_coaltown_intermediate2", "Quarry", MapCoaltown, DifficultyIntermediate, 6, true},
	{11, "mvm_coaltown_advanced", "Ctrl+Alt+Destruction", MapCoaltown, DifficultyAdvanced, 7, true},
	{12, "mvm_coaltown_advanced2", "CPU Slaughter", MapCoaltown, DifficultyAdvanced, 6, true},
	{13, "mvm_coaltown_expert1", "Cataclysm", MapCoaltown, DifficultyExpert, 7, true},
	{14, "mvm_mannworks", "Mann-euvers", MapMannworks, DifficultyNormal, 7, true},
	{15, "mvm_mannworks_intermediate", "Mean Machines", MapMannworks, DifficultyIntermediate, 6, true},
	{16, "mvm_mannworks_intermediate2", "Mannhunt", MapMannworks, DifficultyIntermediate, 6, true},
	{17, "mvm_mannworks_advanced", "Machine Massacre", MapMannworks, DifficultyAdvanced, 7, true},
	{18, "mvm_mannworks_ironman", "Mech Mutilation", MapMannworks, DifficultyAdvanced, 3, true},
	{19, "mvm_mannworks_expert1", "Mannslaughter", MapMannworks, DifficultyExpert, 5, true},
	{20, "mvm_bigrock", "Benign Infiltration", MapBigrock, DifficultyNormal, 6, true},
	{21, "mvm_bigrock_advanced1", "Broken Parts", MapBigrock, DifficultyAdvanced, 7, true},
	{22, "mvm_bigrock_advanced2", "Bone Shaker", MapBigrock, DifficultyAdvanced, 8, true},
	{23, "mvm_mannhattan", "Big Apple Barricade", MapMannhattan, DifficultyIntermediate, 6, false},
	{24, "mvm_mannhattan_advanced1", "Empire Escalation", MapMannhattan, DifficultyAdvanced, 6, false},
	{25, "mvm_mannhattan_advanced2", "Metro Malice", MapMannhattan, DifficultyAdvanced, 6, false},
	{26, "mvm_rottenburg", "Village Vanguard", MapRottenburg, DifficultyIntermediate, 7, true},
	{27, "mvm_rottenburg_advanced1", "Hamlet Hostility", MapRottenburg, DifficultyAdvanced, 7, true},
	{28, "mvm_rottenburg_advanced2", "Bavarian Botbash", MapRottenburg, DifficultyAdvanced, 7, true},
	{29, "mvm_ghost_town_666", "Caliginous Caper", MapGhostTown, DifficultyHaunted, 1, true},
}

var (
	missionsByPopFile = indexMissionsByPopFile()
	missionsByID      = indexMissionsByID()
)

func indexMissionsByPopFile() map[string]Mission {
	byPopFile := make(map[string]Mission, len(Missions))
	for _, m := range Missions {
		byPopFile[m.PopFile] = m
	}
	return byPopFile
}

func indexMissionsByID() map[MissionID]Mission {
	byID := make(map[MissionID]Mission, len(Missions))
	for _, m := range Missions {
		byID[m.ID] = m
	}
	return byID
}

// MissionByID returns the mission an item or a location belongs to.
func MissionByID(id MissionID) (Mission, bool) {
	m, ok := missionsByID[id]
	return m, ok
}

// MissionByPopFile returns the mission the plugin means when it reports an
// objective. The pop file name is the only part of a mission the game knows.
func MissionByPopFile(popFile string) (Mission, bool) {
	m, ok := missionsByPopFile[popFile]
	return m, ok
}

// MissionsByDifficulty returns the missions in one tier, in table order.
func MissionsByDifficulty(d Difficulty) []Mission {
	tier := make([]Mission, 0, len(Missions))
	for _, m := range Missions {
		if m.Difficulty == d {
			tier = append(tier, m)
		}
	}
	return tier
}
