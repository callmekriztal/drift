package palette

import "strings"

// The three built-in palettes. Each is a full day cycle: colors shift with the
// local hour, so the same palette reads as deep night at 2am and soft haze at
// noon, always inside its own family.

// NordNight is the default: blue-hour blues, snow-storm moonlight.
func NordNight() *Palette {
	p := &Palette{Name: "nord-night", Keys: []Key{
		// Deep night
		{Hour: 0, Zenith: "#0E1420", Horizon: "#1D2740", HillFar: "#1A2436", HillNear: "#0A0F1A",
			Water: "#6E7F9A", Moon: "#E9EDF3", Reed: "#080C14", Whisper: "#88C0D0",
			MoonA: 1.0, StarA: 1.0, FlyA: 1.0},
		// Late night, first hint of morning
		{Hour: 4, Zenith: "#111828", Horizon: "#28344F", HillFar: "#1D2839", HillNear: "#0B101B",
			Water: "#71809C", Moon: "#E9EDF3", Reed: "#080C14", Whisper: "#88C0D0",
			MoonA: 0.92, StarA: 0.85, FlyA: 0.9},
		// Dawn: cold rose over the ridge
		{Hour: 6, Zenith: "#33415F", Horizon: "#C08255", HillFar: "#3E4C6A", HillNear: "#1D2739",
			Water: "#8D7F77", Moon: "#E9EDF3", Reed: "#11161F", Whisper: "#81A1C1",
			MoonA: 0.18, StarA: 0.12, FlyA: 0.25},
		// Morning: nord frost light
		{Hour: 8.5, Zenith: "#54749B", Horizon: "#AEBECD", HillFar: "#71839B", HillNear: "#394352",
			Water: "#A9B4C0", Moon: "#E9EDF3", Reed: "#242B36", Whisper: "#4C566A",
			MoonA: 0, StarA: 0, FlyA: 0.05},
		// Midday haze
		{Hour: 13, Zenith: "#6488B2", Horizon: "#C3D0DC", HillFar: "#7F91A8", HillNear: "#3F4956",
			Water: "#B9C4CF", Moon: "#E9EDF3", Reed: "#242B36", Whisper: "#3B4252",
			MoonA: 0, StarA: 0, FlyA: 0},
		// Late afternoon
		{Hour: 16.5, Zenith: "#5D7799", Horizon: "#D2C2B4", HillFar: "#71819B", HillNear: "#394352",
			Water: "#B2AEAE", Moon: "#E9EDF3", Reed: "#1F2630", Whisper: "#434C5E",
			MoonA: 0, StarA: 0, FlyA: 0},
		// Golden hour
		{Hour: 18.6, Zenith: "#4C5D82", Horizon: "#DFA072", HillFar: "#5B687F", HillNear: "#2C3441",
			Water: "#B09684", Moon: "#E9EDF3", Reed: "#161B24", Whisper: "#5E81AC",
			MoonA: 0.08, StarA: 0.05, FlyA: 0.2},
		// Sunset embers
		{Hour: 19.6, Zenith: "#37436A", Horizon: "#CE8462", HillFar: "#414D6A", HillNear: "#20293A",
			Water: "#967E77", Moon: "#E9EDF3", Reed: "#10151E", Whisper: "#88C0D0",
			MoonA: 0.3, StarA: 0.2, FlyA: 0.6},
		// Dusk: stars return
		{Hour: 21, Zenith: "#182238", Horizon: "#3C4A73", HillFar: "#1E2940", HillNear: "#101723",
			Water: "#75819C", Moon: "#E9EDF3", Reed: "#0A0E16", Whisper: "#88C0D0",
			MoonA: 0.85, StarA: 0.65, FlyA: 0.9},
		// Night settles
		{Hour: 22.5, Zenith: "#111827", Horizon: "#25314D", HillFar: "#17202F", HillNear: "#0B101A",
			Water: "#70809B", Moon: "#E9EDF3", Reed: "#080C14", Whisper: "#88C0D0",
			MoonA: 1.0, StarA: 1.0, FlyA: 1.0},
	}}
	p.normalize()
	return p
}

// RoséPineDawn: warm paper tones, rose light, pine shadows.
func RoséPineDawn() *Palette {
	p := &Palette{Name: "rose-pine-dawn", Keys: []Key{
		{Hour: 0, Zenith: "#3F3B58", Horizon: "#6E5F80", HillFar: "#555074", HillNear: "#2F2B45",
			Water: "#5E5880", Moon: "#F5E9DF", Reed: "#221E36", Whisper: "#B4637A",
			MoonA: 0.9, StarA: 0.9, FlyA: 1.0},
		{Hour: 4, Zenith: "#46415F", Horizon: "#77678A", HillFar: "#5B5578", HillNear: "#332E4C",
			Water: "#635D84", Moon: "#F5E9DF", Reed: "#221E36", Whisper: "#B4637A",
			MoonA: 0.85, StarA: 0.75, FlyA: 0.9},
		// Dawn bloom
		{Hour: 6, Zenith: "#8A7BA0", Horizon: "#E8A668", HillFar: "#9A8AAC", HillNear: "#4E4670",
			Water: "#9C8AA2", Moon: "#F5E9DF", Reed: "#332E4C", Whisper: "#286983",
			MoonA: 0.15, StarA: 0.1, FlyA: 0.35},
		// Morning: rosy paper
		{Hour: 9, Zenith: "#BFB8D4", Horizon: "#F5E3D0", HillFar: "#B0A3C0", HillNear: "#655E86",
			Water: "#CBBFC8", Moon: "#F5E9DF", Reed: "#4B4468", Whisper: "#575279",
			MoonA: 0, StarA: 0, FlyA: 0.05},
		// Noon
		{Hour: 13, Zenith: "#C9C4DE", Horizon: "#FBF0E4", HillFar: "#BDB4CE", HillNear: "#6B6390",
			Water: "#D8CCD4", Moon: "#F5E9DF", Reed: "#4B4468", Whisper: "#575279",
			MoonA: 0, StarA: 0, FlyA: 0},
		// Afternoon
		{Hour: 16.5, Zenith: "#B4A8CC", Horizon: "#F3DCC6", HillFar: "#AC9EC0", HillNear: "#615983",
			Water: "#CDBEBE", Moon: "#F5E9DF", Reed: "#453E62", Whisper: "#575279",
			MoonA: 0, StarA: 0, FlyA: 0},
		// Golden hour
		{Hour: 18.6, Zenith: "#9C86B4", Horizon: "#EA9D34", HillFar: "#967FAE", HillNear: "#514A72",
			Water: "#C0A08E", Moon: "#F5E9DF", Reed: "#3A3454", Whisper: "#3E5F76",
			MoonA: 0.1, StarA: 0.05, FlyA: 0.2},
		// Dusk rose
		{Hour: 20.5, Zenith: "#6E6390", Horizon: "#CE8278", HillFar: "#766C97", HillNear: "#3D3760",
			Water: "#8D7A94", Moon: "#F5E9DF", Reed: "#2A2545", Whisper: "#8B7BA6",
			MoonA: 0.55, StarA: 0.4, FlyA: 0.85},
		// Night
		{Hour: 22, Zenith: "#45415F", Horizon: "#6E5F80", HillFar: "#5A5474", HillNear: "#35314B",
			Water: "#635D84", Moon: "#F5E9DF", Reed: "#221E36", Whisper: "#B4637A",
			MoonA: 0.95, StarA: 0.95, FlyA: 1.0},
	}}
	p.normalize()
	return p
}

// LavenderHaze: dreamy purple-pink, the softest of the three.
func LavenderHaze() *Palette {
	p := &Palette{Name: "lavender-haze", Keys: []Key{
		{Hour: 0, Zenith: "#1D1730", Horizon: "#43325C", HillFar: "#2B2145", HillNear: "#150F26",
			Water: "#4C3D66", Moon: "#F3EAF9", Reed: "#100B1C", Whisper: "#C9A9E0",
			MoonA: 1.0, StarA: 1.0, FlyA: 1.0},
		{Hour: 4, Zenith: "#221A38", Horizon: "#4C3966", HillFar: "#312650", HillNear: "#18122A",
			Water: "#544272", Moon: "#F3EAF9", Reed: "#100B1C", Whisper: "#C9A9E0",
			MoonA: 0.95, StarA: 0.85, FlyA: 0.95},
		// Pink dawn
		{Hour: 6.5, Zenith: "#6E5590", Horizon: "#EFA8B8", HillFar: "#8468A4", HillNear: "#433463",
			Water: "#A4779B", Moon: "#F3EAF9", Reed: "#2A2145", Whisper: "#9D7BB8",
			MoonA: 0.15, StarA: 0.08, FlyA: 0.3},
		// Day lavender
		{Hour: 10, Zenith: "#9A8CC0", Horizon: "#E6C9E0", HillFar: "#8F7FB4", HillNear: "#4A3D6B",
			Water: "#B9A6CE", Moon: "#F3EAF9", Reed: "#4A3D6B", Whisper: "#6B5A8E",
			MoonA: 0, StarA: 0, FlyA: 0.05},
		// Noon
		{Hour: 14, Zenith: "#A795CC", Horizon: "#EFD8EC", HillFar: "#9A88BE", HillNear: "#4F4172",
			Water: "#C4B2D8", Moon: "#F3EAF9", Reed: "#4A3D6B", Whisper: "#6B5A8E",
			MoonA: 0, StarA: 0, FlyA: 0},
		// Pink-gold afternoon
		{Hour: 17.5, Zenith: "#8A6BA8", Horizon: "#F0A8A0", HillFar: "#84639F", HillNear: "#443265",
			Water: "#B083A4", Moon: "#F3EAF9", Reed: "#33254E", Whisper: "#8A6BA8",
			MoonA: 0.05, StarA: 0, FlyA: 0.2},
		// Violet dusk
		{Hour: 20, Zenith: "#4A3870", Horizon: "#8A5E90", HillFar: "#564279", HillNear: "#2A1F44",
			Water: "#664F86", Moon: "#F3EAF9", Reed: "#1E1633", Whisper: "#C9A9E0",
			MoonA: 0.7, StarA: 0.5, FlyA: 0.95},
		// Night
		{Hour: 22, Zenith: "#1D1730", Horizon: "#43325C", HillFar: "#2B2145", HillNear: "#150F26",
			Water: "#4C3D66", Moon: "#F3EAF9", Reed: "#100B1C", Whisper: "#C9A9E0",
			MoonA: 1.0, StarA: 1.0, FlyA: 1.0},
	}}
	p.normalize()
	return p
}

// All returns the built-in palettes in cycle order.
func All() []*Palette {
	return []*Palette{NordNight(), RoséPineDawn(), LavenderHaze()}
}

// Get finds a palette by (case-insensitive) name or common alias.
func Get(name string) *Palette {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, " ", "-")
	for _, p := range All() {
		if p.Name == name {
			return p
		}
	}
	switch name {
	case "nord", "night":
		return NordNight()
	case "rose", "rose-pine", "rosey-pine", "dawn":
		return RoséPineDawn()
	case "lavender", "haze":
		return LavenderHaze()
	}
	return nil
}
