package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func main() {
	// Path to the specific Go program
	goProgramPath := filepath.Join("greyscale", "greyscale.go")

	// Number of times to run the program
	numRuns := 110
	cpuTimes := make([]float64, numRuns-10)

	// Run the Go program and measure CPU time
	for i := 0; i < numRuns; i++ {
		startTime := time.Now()

		// Run the Go program using exec.Command
		cmd := exec.Command("go", "run", goProgramPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()

		endTime := time.Now()

		// Calculate CPU time in seconds
		cpuTime := endTime.Sub(startTime).Seconds()
		// Disconsidering the first 10 results
		if i > 9 {
			cpuTimes[i-10] = cpuTime
		}
		fmt.Println("Começando Interação: ", i, cpuTime)
	}

	// Calculate the average CPU time
	averageCPUTime := calculateAverage(cpuTimes)
	// Output the results to the end of the file in the "exercicio1" subfolder
	outputFilePath := filepath.Join("greyscale", "average_cpu_time.txt")
	outputFile, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening output file:", err)
		return
	}
	defer outputFile.Close()

	_, err = fmt.Fprintf(outputFile, "Average CPU Time: %f seconds\n", averageCPUTime)
	if err != nil {
		fmt.Println("Error writing to output file:", err)
		return
	}

	fmt.Printf("Average CPU Time: %f seconds\n", averageCPUTime)
	fmt.Printf("Results saved to %s\n", outputFilePath)
}

// calculateAverage calculates the average of a slice of float64 values
func calculateAverage(values []float64) float64 {
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}
