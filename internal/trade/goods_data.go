package trade

// The Random Trade Goods chart (Book 2 pp.218-219), transcribed from the
// rendered page images. Each trade-classification column has six blocks selected
// by a first 1D; a second 1D selects one of six goods within the block. A block
// typed "Imbalances" holds trade-class codes to recurse on instead of goods.

// goodsBlock is one first-1D block of a column: a type category and its six
// goods (or, for an Imbalances block, six trade-class codes).
type goodsBlock struct {
	Type  string
	Goods [6]string
}

var tradeGoodsColumns = map[string][6]goodsBlock{
	"Ag-1": {
		{"Raws", [6]string{"Bulk Protein", "Bulk Carbs", "Bulk Fats", "Bulk Pharma", "Livestock", "Seedstock"}},
		{"Consumables", [6]string{"Flavored Waters", "Wines", "Juices", "Nectars", "Decoctions", "Drinkable Lymphs"}},
		{"Pharma", [6]string{"Health Foods", "Nutraceuticals", "Fast Drug", "Painkillers", "Antiseptic", "Antibiotics"}},
		{"Novelties", [6]string{"Incenses", "Iridescents", "Photonics", "Pigments", "Noisemakers", "Soundmakers"}},
		{"Rares", [6]string{"Fine Furs", "Meat Delicacies", "Fruit Delicacies", "Candies", "Textiles", "Exotic Sauces"}},
		{"Imbalances", [6]string{"As", "De", "Fl", "Ic", "Na", "In"}},
	},
	"Ag-2": {
		{"Raws", [6]string{"Bulk Woods", "Bulk Pelts", "Bulk Herbs", "Bulk Spices", "Bulk Nitrates", "Foodstuffs"}},
		{"Consumables", [6]string{"Flowers", "Aromatics", "Pheromones", "Secretions", "Adhesives", "Novel Flavorings"}},
		{"Pharma", [6]string{"Antifungals", "Antivirals", "Panacea", "Pseudomones", "Anagathics", "Slow Drug"}},
		{"Novelties", [6]string{"Strange Seeds", "Motile Plants", "Reactive Plants", "Reactive Woods", "IR Emitters", "Lek Emitters"}},
		{"Rares", [6]string{"Spices", "Organic Gems", "Flavorings", "Aged Meats", "Fermented Fluids", "Fine Aromatics"}},
		{"Imbalances", [6]string{"Po", "Ri", "Va", "Ic", "Na", "In"}},
	},
	"As": {
		{"Raws", [6]string{"Bulk Nitrates", "Bulk Carbon", "Bulk Iron", "Bulk Copper", "Radioactive Ores", "Bulk Ices"}},
		{"Samples", [6]string{"Ores", "Ices", "Carbons", "Metals", "Uranium", "Chelates"}},
		{"Valuta", [6]string{"Platinum", "Gold", "Gallium", "Silver", "Thorium", "Radium"}},
		{"Novelties", [6]string{"Unusual Rocks", "Fused Metals", "Strange Crystals", "Fine Dusts", "Magnetics", "Light-Sensitives"}},
		{"Rares", [6]string{"Gemstones", "Alloys", "Iridium Sponge", "Lanthanum", "Isotopes", "Anti-Matter"}},
		{"Imbalances", [6]string{"Ag", "De", "Na", "Po", "Ri", "Ic"}},
	},
	"De": {
		{"Raws", [6]string{"Bulk Nitrates", "Bulk Minerals", "Bulk Abrasives", "Bulk Particulates", "Exotic Fauna", "Exotic Flora"}},
		{"Samples", [6]string{"Archeologicals", "Fauna", "Flora", "Minerals", "Ephemerals", "Polymers"}},
		{"Pharma", [6]string{"Stimulants", "Bulk Herbs", "Palliatives", "Pheromones", "Antibiotics", "Combat Drug"}},
		{"Novelties", [6]string{"Envirosuits", "Reclamation Suits", "Navigators", "Dupe Masterpieces", "ShimmerCloth", "ANIFX Blocker"}},
		{"Rares", [6]string{"Excretions", "Flavorings", "Nectars", "Pelts", "ANIFX Dyes", "Seedstock"}},
		{"Uniques", [6]string{"Pheromones", "Artifacts", "Sparx", "Repulsant", "Dominants", "Fossils"}},
	},
	"Fl": {
		{"Raws", [6]string{"Bulk Carbon", "Bulk Petros", "Bulk Precipitates", "Exotic Fluids", "Organic Polymers", "Bulk Synthetics"}},
		{"Samples", [6]string{"Archeologicals", "Fauna", "Flora", "Germanes", "Flill", "Chelates"}},
		{"Pharma", [6]string{"Antifungals", "Antivirals", "Palliatives", "Counter-prions", "Antibiotics", "Cold Sleep Pills"}},
		{"Novelties", [6]string{"Silanes", "Lek Emitters", "Aware Blockers", "Soothants", "Self-Solving Puzzles", "Fluidic Timepieces"}},
		{"Rares", [6]string{"Flavorings", "Unusual Fluids", "Encapsulants", "Insidiants", "Corrosives", "Exotic Aromatics"}},
		{"Imbalances", [6]string{"In", "Ri", "Ic", "Na", "Ag", "Po"}},
	},
	"Ic": {
		{"Raws", [6]string{"Bulk Ices", "Bulk Precipitates", "Bulk Ephemerals", "Exotic Flora", "Bulk Gases", "Bulk Oxygen"}},
		{"Samples", [6]string{"Archeologicals", "Fauna", "Flora", "Minerals", "Luminescents", "Polymers"}},
		{"Pharma", [6]string{"Antifungals", "Antivirals", "Palliatives", "Restoratives", "Antibiotics", "Antiseptics"}},
		{"Novelties", [6]string{"Heat Pumps", "Mag Emitters", "Percept Blockers", "Silanes", "Cold Light Blocks", "VHDUS Blocker"}},
		{"Rares", [6]string{"Unusual Ices", "Cryo Alloys", "Rare Minerals", "Unusual Fluids", "Cryogems", "VHDUS Dyes"}},
		{"Uniques", [6]string{"Fossils", "Cryogems", "Vision Suppressant", "Fission Suppressant", "Wafers", "Cold Sleep Pills"}},
	},
	"Na": {
		{"Raws", [6]string{"Bulk Abrasives", "Bulk Gases", "Bulk Minerals", "Bulk Precipitates", "Exotic Fauna", "Exotic Flora"}},
		{"Samples", [6]string{"Archeologicals", "Fauna", "Flora", "Minerals", "Ephemerals", "Polymers"}},
		{"Novelties", [6]string{"Branded Tools", "Drinkable Lymphs", "Strange Seeds", "Pattern Creators", "Pigments", "Warm Leather"}},
		{"Rares", [6]string{"Hummingsand", "Masterpieces", "Fine Carpets", "Isotopes", "Pelts", "Seedstock"}},
		{"Uniques", [6]string{"Masterpieces", "Unusual Rocks", "Artifacts", "Non-Fossil Carcasses", "Replicating Clays", "ANIFX Emitter"}},
		{"Imbalances", [6]string{"Ag", "Ri", "In", "Ic", "De", "Fl"}},
	},
	"In": {
		{"Manufactureds", [6]string{"Electronics", "Photonics", "Magnetics", "Fluidics", "Polymers", "Gravitics"}},
		{"Scrap/Waste", [6]string{"Obsoletes", "Used Goods", "Reparables", "Radioactives", "Metals", "Sludges"}},
		{"Manufactureds", [6]string{"Biologics", "Mechanicals", "Textiles", "Weapons", "Armor", "Robots"}},
		{"Pharma", [6]string{"Nostrums", "Restoratives", "Palliatives", "Chelates", "Antidotes", "Antitoxins"}},
		{"Data", [6]string{"Software", "Databases", "Expert Systems", "Upgrades", "Backups", "Raw Sensings"}},
		{"Consumables", [6]string{"Disposables", "Respirators", "Filter Masks", "Combination", "Parts", "Improvements"}},
	},
	"Po": {
		{"Raws", [6]string{"Bulk Nutrients", "Bulk Fibers", "Bulk Organics", "Bulk Minerals", "Bulk Textiles", "Exotic Flora"}},
		{"Entertainments", [6]string{"Art", "Recordings", "Writings", "Tactiles", "Osmancies", "Wafers"}},
		{"Novelties", [6]string{"Strange Crystals", "Strange Seeds", "Pigments", "Emotion Lighting", "Silanes", "Flora"}},
		{"Rares", [6]string{"Gemstones", "Antiques", "Collectibles", "Allotropes", "Spices", "Seedstock"}},
		{"Uniques", [6]string{"Masterpieces", "Exotic Flora", "Antiques", "Incomprehensibles", "Fossils", "VHDUS Emitter"}},
		{"Imbalances", [6]string{"In", "Ri", "Fl", "Ic", "Ag", "Va"}},
	},
	"Ri": {
		{"Raws", [6]string{"Bulk Foodstuffs", "Bulk Protein", "Bulk Carbs", "Bulk Fats", "Exotic Flora", "Exotic Fauna"}},
		{"Novelties", [6]string{"Echostones", "Self-Defenders", "Attractants", "Sophont Cuisine", "Sophont Hats", "Variable Tattoos"}},
		{"Consumables", [6]string{"Branded Foods", "Branded Drinks", "Branded Clothes", "Flavored Drinks", "Flowers", "Music"}},
		{"Rares", [6]string{"Delicacies", "Spices", "Tisanes", "Nectars", "Pelts", "Variable Tattoos"}},
		{"Uniques", [6]string{"Antique Art", "Masterpieces", "Artifacts", "Fine Art", "Meson Barriers", "Famous Wafers"}},
		{"Entertainments", [6]string{"Edutainments", "Recordings", "Writings", "Tactiles", "Osmancies", "Wafers"}},
	},
	"Va": {
		{"Raws", [6]string{"Bulk Dusts", "Bulk Minerals", "Bulk Metals", "Radioactive Ores", "Bulk Particulates", "Ephemerals"}},
		{"Novelties", [6]string{"Branded Vacc Suits", "Awareness Pinger", "Strange Seeds", "Pigments", "Unusual Minerals", "Exotic Crystals"}},
		{"Consumables", [6]string{"Branded Oxygen", "Vacc Suit Scents", "Vacc Suit Patches", "Branded Tools", "Holo-Companions", "Flavored Air"}},
		{"Rares", [6]string{"Vacc Gems", "Unusual Dusts", "Insulants", "Crafted Devices", "Rare Minerals", "Catalysts"}},
		{"Samples", [6]string{"Archeologicals", "Fauna", "Flora", "Minerals", "Ephemerals", "Polymers"}},
		{"Scrap/Waste", [6]string{"Obsoletes", "Used Goods", "Reparables", "Plutonium", "Metals", "Sludges"}},
	},
	"CpCsCx": {
		{"Data", [6]string{"Software", "Expert Systems", "Databases", "Upgrades", "Backups", "Raw Sensings"}},
		{"Novelties", [6]string{"Incenses", "Contemplatives", "Cold Welders", "Polymer Sheets", "Hats", "Skin Tones"}},
		{"Consumables", [6]string{"Branded Clothes", "Branded Devices", "Flavored Drinks", "Flavorings", "Decorations", "Group Symbols"}},
		{"Rares", [6]string{"Monumental Art", "Holo Sculpture", "Collectible Books", "Jewelry", "Museum Items", "Monumental Art"}},
		{"Valuta", [6]string{"Coinage", "Currency", "Money Cards", "Gold", "Silver", "Platinum"}},
		{"Red Tape", [6]string{"Regulations", "Synchronizations", "Expert Systems", "Educationals", "Mandates", "Accountings"}},
	},
}

// tradeGoodsDetailOrder lists the detail-bearing trade classes in the order the
// p.219 Trade Good Detail chart prints them. The book only says "select one" when
// several apply, so tradeGoodsDetail takes the first in this order — a fixed rule,
// so a world's cargo always describes the same way.
var tradeGoodsDetailOrder = []string{
	"As", "Ba", "De", "Di", "Fl", "Ga", "He", "Hi", "Ic", "Ni", "Po", "Ri", "Va", "Wa",
}

// tradeGoodsDetailLabel is the Trade Good Detail prefix each trade class confers
// (Book 2 p.219). Hi's "Processed" is omitted for goods out of the Industrial
// column and Va's "Exotic" for goods out of the Asteroid column, where the label
// would restate the goods' own origin (handled in tradeGoodsDetail).
var tradeGoodsDetailLabel = map[string]string{
	"As": "Strange", "Ba": "Gathered", "De": "Mineral", "Di": "Artifact",
	"Fl": "Unusual", "Ga": "Premium", "He": "Strange", "Hi": "Processed",
	"Ic": "Cryo", "Ni": "Unprocessed", "Po": "Obscure", "Ri": "Quality",
	"Va": "Exotic", "Wa": "Infused",
}

// goodsColumnEligible is the set of trade classes that select a Random Trade
// Goods column (Book 2 p.218 "Which Trade Classification To Use").
var goodsColumnEligible = map[string]bool{
	"Ag": true, "As": true, "De": true, "Fa": true, "Fl": true, "Ga": true,
	"Ic": true, "In": true, "Na": true, "Po": true, "Ri": true, "Va": true,
	"Cp": true, "Cs": true, "Cx": true,
}
