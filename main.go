package main

import (
  "bufio"
  "fmt"
  "os"
  "strings"
  
  fr "simplex/fraction"
  tb "simplex/tableau"
)

func main() {
  reader := bufio.NewReader(os.Stdin)
  var st tb.Tableau
  var constraintCount, variableCount int
  var problemType string

  fmt.Print("Are you solving a maximization or minimization problem? (max/min): ")
  problemType, _ = reader.ReadString('\n')
  problemType = strings.TrimSpace(problemType)
  
  isMaximization := true
  if strings.ToLower(problemType) == "min" {
    isMaximization = false
  }

  fmt.Print("Enter number of constraints: ")
  fmt.Scan(&constraintCount)

  fmt.Print("Enter number of variables: ")
  fmt.Scan(&variableCount)
  
  // Initialize the tableau (constraints + objective row, variables + constant column)
  st.Init(constraintCount + 1, variableCount + 1)
  st.SetMaximization(isMaximization)
  
  fmt.Println("\nEnter column names (e.g., x1, x2, etc.):")
  for j := 0; j < variableCount; j++ {
    fmt.Printf("Column %d name: ", j+1)
    var name string
    fmt.Scan(&name)
    st.ColNames[j] = name
  }
  st.ColNames[variableCount] = "const" // Last column is constants
  
  fmt.Println("\nEnter row names (e.g., s1, s2, etc.):")
  for i := 0; i < constraintCount; i++ {
    fmt.Printf("Row %d name: ", i+1)
    var name string
    fmt.Scan(&name)
    st.RowNames[i] = name
  }
  st.RowNames[constraintCount] = "F" // Last row is objective function
  
  fmt.Println("\nNow enter the tableau coefficients:")
  for i := 0; i < constraintCount+1; i++ {
    fmt.Printf("Enter values for row %s:\n", st.RowNames[i])
    for j := 0; j < variableCount+1; j++ {
      fmt.Printf("  Coefficient for %s: ", st.ColNames[j])
      fr.Read(&st.Table[i][j])
    }
  }

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
  
  // Print variable values
  for varName, value := range solution {
    if varName != "objective" {
      fmt.Printf("%s = ", varName)
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
