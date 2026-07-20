# trade

The Trade & Commerce pricing engine (Book 2 pp. 209-221). Speculative cargo is bought at a source
world for its `Cost` (Cr3,000 base + per-value-class cost mods + Cr100/TL) and sold at a market
world for a fraction/multiple of its `Price` (Cr5,000 base + source→market match mods, scaled
10%/TL-difference), realized through the `ActualValue` table (Flux→40–400%, with the capped Broker
DM). Pure int-Cr, no dice in the value math; `CargoID` renders the p.221 identity (e.g. `8-De Hi In
Na Po Cr1,800`). Golden-locked to the Free Trader Beowulf journey.

`shipping.go` adds premium passage pricing, passenger/freight availability + rates, and the Broker
table (starport gating + commission + `NetSale`). `goods.go`/`goods_data.go` add the Random Trade
Goods chart (12 TC-keyed columns, 1D type→1D good, Imbalance recursion, Trade Good Detail prefix;
golden-locked to the Zivije/Knorbes examples). `contracts.go` adds Trader estimation
(`EstimateActualValue`), the OTO/STS and accelerated-delivery surcharges, and the long-term
mail-contract bid table. #21 is complete.
