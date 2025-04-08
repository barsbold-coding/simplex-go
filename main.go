package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	
	fr "simplex/fraction"
	"simplex/parser"
	tb "simplex/tableau"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	var problemType string

	fmt.Print("Are you solving a maximization or minimization problem? (max/min): ")
	problemType, _ = reader.ReadString('\n')
	problemType = strings.TrimSpace(problemType)
	
	isMaximization := true
	if strings.ToLower(problemType) == "min" {
		isMaximization = false
	}

	fmt.Print("Enter the objective function (e.g., '2x1 + 3x2'): ")
	objectiveStr, _ := reader.ReadString('\n')
	objectiveStr = strings.TrimSpace(objectiveStr)

	fmt.Print("Enter the number of constraints: ")
	var constraintCount int
	fmt.Scan(&constraintCount)
	reader.ReadString('\n') // Consume newline

	constraintStrs := make([]string, constraintCount)
	fmt.Println("\nEnter your constraints (e.g., '3x1 + 2x2 <= 6'):")
	for i := 0; i < constraintCount; i++ {
		fmt.Printf("Constraint %d: ", i+1)
		constraintStrs[i], _ = reader.ReadString('\n')
		constraintStrs[i] = strings.TrimSpace(constraintStrs[i])
	}

	problem, err := parser.ParseProblem(objectiveStr, constraintStrs, isMaximization)
	if err != nil {
		fmt.Printf("Error parsing problem: %v\n", err)
		return
	}

	fmt.Println("\nParsed problem successfully...")
	fmt.Printf("Objective: %s\n", objectiveStr)
	for i, constraint := range constraintStrs {
		fmt.Printf("Constraint %d: %s\n", i+1, constraint)
	}

	// Convert the problem to tableau format
	st := parser.ConvertToTableau(problem)

	fmt.Println("\nInitial Tableau:")
	tb.Print(&st)
	
	if !st.IsFeasible() {
		fmt.Println("Warning: Initial tableau is not feasible (contains negative RHS values)")

		if !st.MakeFeasible() {
			fmt.Println("Failed to find feasible solution. Problem may be infeasible.")
			return
		}
	}
	
	iteration := 1
	for {
		if st.IsOptimal() {
			fmt.Println("Optimal solution reached!")
			break
		}
		
		fmt.Printf("\n--- Iteration %d ---\n", iteration)
		r, s := st.Pivot()
		if !tb.IsPivotValid(r, s) { 
			fmt.Println("No valid pivot found. Solution may be unbounded.")
			break 
		}

		fmt.Printf("Pivoting on element at row %d, column %d (intersection of %s and %s)\n", 
				r, s, st.RowNames[r], st.ColNames[s])
		b := st.Transform(r, s)
		st = b.Copy()
		
		tb.Print(&st)
		iteration++
		
		// Safety check to prevent infinite loops
		if iteration > 100 {
			fmt.Println("Warning: Maximum iterations reached. Process stopped.")
			break
		}
	}

	// Print final solution
	fmt.Println("\nFinal Tableau:")
	tb.Print(&st)
	
	solution := st.GetSolution()
	fmt.Println("\nSolution:")
	
	// Print variable values (focus on the original decision variables)
	for varName, value := range solution {
		// Handle both possible tableau formats (variable as row or column)
		cleanName := strings.TrimPrefix(varName, "-")
		if varName != "objective" && !strings.HasPrefix(varName, "s") {
			fmt.Printf("%s = ", cleanName)
			fr.Print(&value, 0)
			fmt.Println()
		}
	}
	
	// Print objective value
	fmt.Print("\nObjective value = ")
	objectiveValue := solution["objective"] 
	if !isMaximization {
		// For minimization problems, we typically negate the final objective value
		objectiveValue = fr.Neg(objectiveValue)
	}
	fr.Print(&objectiveValue, 0)
	fmt.Println()
}
