package main

import (
	"fmt"
	"math"
	"strconv"
	"syscall/js"
)

type SimulationResult struct {
	deadBirds []int
	totalCost float64
}

func main() {
	c := make(chan struct{}, 0)
	js.Global().Set("runSimulation", js.FuncOf(runSimulation))
	<-c
}

func runSimulation(this js.Value, args []js.Value) interface{} {
	strategy := args[0].String()
	mortalityRateStr := args[1].String()
	mortalityRate, _ := strconv.ParseFloat(mortalityRateStr, 64)
	mortalityRate = mortalityRate / 100.0
	initialPopulation := 600000
	infected := 1000
	r0 := 3.0
	cycles := 3
	cullingThreshold := 1000
	cullingAmount := 200000

	var result SimulationResult
	if strategy == "culling" {
		result = simulateWithCulling(initialPopulation, infected, r0, cycles, cullingThreshold, cullingAmount, mortalityRate)
	} else {
		result = simulateWithoutCulling(initialPopulation, infected, r0, cycles, mortalityRate)
	}

	// Update results text
	resultText := fmt.Sprintf("Total cost: $%.2f", result.totalCost)
	js.Global().Get("document").Call("getElementById", "results").Set("innerHTML", resultText)

	// Draw graph
	drawGraph(result.deadBirds, strategy)

	return nil
}

func simulateWithCulling(population, infected int, r0 float64, cycles, threshold, cullingAmount int, mortalityRate float64) SimulationResult {
	deadBirds := make([]int, cycles)
	totalCost := 0.0

	for i := 0; i < cycles; i++ {
		// Each cycle results in cullingAmount (200,000) dead birds
		deadBirds[i] = cullingAmount

		// Add cost for this cycle's dead birds ($3 per bird)
		totalCost += float64(cullingAmount) * 3
	}

	return SimulationResult{
		deadBirds: deadBirds,
		totalCost: totalCost,
	}
}

func simulateWithoutCulling(population, infected int, r0 float64, cycles int, mortalityRate float64) SimulationResult {
	deadBirds := make([]int, cycles)
	totalDead := 0
	totalCost := 0.0
	daysPerCycle := 120
	beta := r0 / 7.0   // Transmission rate assuming 7 day infectious period
	gamma := 1.0 / 7.0 // Recovery rate (1/infectious period)

	for i := 0; i < cycles; i++ {
		// Restore population at start of cycle and add cost
		if i > 0 {
			totalCost += float64(totalDead) * 3
			totalDead = 0
		}

		// Start new outbreak each cycle with restored population
		S := float64(population)
		I := float64(infected)
		R := 0.0

		// Run SIR model for this cycle
		for day := 0; day < daysPerCycle; day++ {
			// SIR differential equations
			dS := -beta * S * I / float64(population)
			dI := beta*S*I/float64(population) - gamma*I
			dR := gamma * I

			S += dS
			I += dI
			R += dR
		}

		// Apply mortality rate to recovered birds
		actualDeaths := int(R * mortalityRate)
		deadBirds[i] = totalDead + actualDeaths
		totalDead += actualDeaths
	}

	// Add cost for final cycle's dead birds
	totalCost += float64(totalDead) * 3

	return SimulationResult{
		deadBirds: deadBirds,
		totalCost: totalCost,
	}
}

func drawGraph(data []int, strategy string) {
	canvas := js.Global().Get("document").Call("getElementById", "chart")
	ctx := canvas.Call("getContext", "2d")

	// Clear canvas
	// Get canvas dimensions from actual element size
	width := canvas.Get("width").Float()
	height := canvas.Get("height").Float()

	ctx.Call("clearRect", 0, 0, width, height)

	// Set up graph parameters
	padding := width * 0.06 // Make padding relative to canvas size
	maxValue := 0

	for _, v := range data {
		if v > maxValue {
			maxValue = v
		}
	}

	// Draw axes
	ctx.Set("strokeStyle", "#000")
	ctx.Call("beginPath")
	ctx.Call("moveTo", padding, padding)
	ctx.Call("lineTo", padding, height-padding)
	ctx.Call("lineTo", width-padding, height-padding)
	ctx.Call("stroke")

	// Draw bars
	barWidth := (width - 2*padding) / float64(len(data)) * 0.8 // 80% of available space
	spacing := (width - 2*padding) / float64(len(data)) * 0.2  // 20% spacing
	yScale := (height - 2*padding) / float64(maxValue)

	for i, v := range data {
		x := padding + float64(i)*(barWidth+spacing)
		y := height - padding - float64(v)*yScale
		barHeight := float64(v) * yScale

		ctx.Set("fillStyle", "#00f")
		ctx.Call("fillRect", x, y, barWidth, barHeight)

		// Add value label on top of each bar
		ctx.Set("fillStyle", "#000")
		ctx.Set("font", "12px Arial")
		ctx.Set("textAlign", "center")
		ctx.Call("fillText", fmt.Sprintf("$%s", humanize(v)), x+barWidth/2, y-5)
	}

	// Add labels
	ctx.Set("fillStyle", "#000")
	ctx.Set("font", "12px Arial")
	ctx.Call("fillText", "Cycles", width/2, height-10)
	ctx.Call("save")
	ctx.Call("translate", 15, height/2)
	ctx.Call("rotate", -math.Pi/2)
	ctx.Call("fillText", "Dead Birds", 0, 0)
	ctx.Call("restore")
}

// humanize formats large numbers with comma separators
func humanize(n int) string {
	return strconv.FormatInt(int64(n), 10)
}
