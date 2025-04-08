package tableau

import (
  "fmt"
  fr "simplex/fraction"
)

type Tableau struct {
  Table [][]fr.Fraction
  dirtX []bool
  dirtY []bool
  RowNames []string  // For slack variables (s1, s2, ..., F)
  ColNames []string  // For decision variables (-x1, -x2, ..., const)
  IsMaximization bool // To track if we're maximizing or minimizing
}

func (t *Tableau) Copy() Tableau {
  copyTable := make([][]fr.Fraction, len(t.Table))
  for i := range t.Table {
    copyTable[i] = make([]fr.Fraction, len(t.Table[i]))
    copy(copyTable[i], t.Table[i])
  }

  copyDirtX := make([]bool, len(t.dirtX))
  copy(copyDirtX, t.dirtX)

  copyDirtY := make([]bool, len(t.dirtY))
  copy(copyDirtY, t.dirtY)

  copyRowNames := make([]string, len(t.RowNames))
  copy(copyRowNames, t.RowNames)

  copyColNames := make([]string, len(t.ColNames))
  copy(copyColNames, t.ColNames)

  return Tableau{
    Table:          copyTable,
    dirtX:          copyDirtX,
    dirtY:          copyDirtY,
    RowNames:       copyRowNames,
    ColNames:       copyColNames,
    IsMaximization: t.IsMaximization,
  }
}

func IsPivotValid(r, s int) bool {
  return r >= 0 && s >= 0
}

func (t *Tableau) Init(rows, cols int) {
  // Initialize the table without variable rows/columns
  t.Table = make([][]fr.Fraction, rows)
  for i := range t.Table {
    t.Table[i] = make([]fr.Fraction, cols)
  }

  // Initialize tracking arrays
  t.dirtX = make([]bool, cols)
  t.dirtY = make([]bool, rows)
  
  // Initialize variable names
  t.RowNames = make([]string, rows)
  t.ColNames = make([]string, cols)
  
  // Default row names
  for i := 0; i < rows-1; i++ {
    t.RowNames[i] = fmt.Sprintf("s%d", i+1)
  }
  t.RowNames[rows-1] = "F" // Last row is objective function
  
  // Default column names
  for j := 0; j < cols-1; j++ {
    t.ColNames[j] = fmt.Sprintf("-x%d", j+1)
  }
  t.ColNames[cols-1] = "const" // Last column is constants
  
  // Default to maximization
  t.IsMaximization = true
}

func (t *Tableau) isValidCell(i, j int) bool {
  return !t.dirtX[j] && !t.dirtY[i] && t.Table[i][j].N != 0
}

// Improved pivot selection based on optimization criteria
func (t *Tableau) Pivot() (int, int) {
  m := len(t.Table)    // Number of rows
  n := len(t.Table[0]) // Number of columns
  
  // For maximization: find most negative coefficient in objective function row
  // For minimization: find most positive coefficient in objective function row
  s := -1
  pivotValue := fr.Fraction{N: 0, D: 1}
  
  for j := 0; j < n-1; j++ { // Skip last column (constant)
    if !t.dirtX[j] {
      if t.IsMaximization && t.Table[m-1][j].N < 0 {
        // For maximization, find most negative coefficient
        if s == -1 || t.Table[m-1][j].N < pivotValue.N {
          s = j
          pivotValue = t.Table[m-1][j]
        }
      } else if !t.IsMaximization && t.Table[m-1][j].N > 0 {
        // For minimization, find most positive coefficient
        if s == -1 || t.Table[m-1][j].N > pivotValue.N {
          s = j
          pivotValue = t.Table[m-1][j]
        }
      }
    }
  }
  
  if s == -1 {
    // No suitable entering variable found - optimal solution reached
    return -1, -1
  }
  
  // Find row with minimum ratio test (smallest positive ratio)
  r := -1
  minRatio := fr.Fraction{N: 0, D: 0} // Initialize with "infinity"
  
  for i := 0; i < m-1; i++ { // Skip objective function row
    if !t.dirtY[i] && t.Table[i][s].N > 0 {
      ratio := fr.Div(t.Table[i][n-1], t.Table[i][s]) // const / coefficient
      if r == -1 || (ratio.N > 0 && (minRatio.N <= 0 || 
         (ratio.N * minRatio.D < minRatio.N * ratio.D))) {
        r = i
        minRatio = ratio
      }
    }
  }
  
  if r == -1 {
    // No limiting constraint - unbounded solution
    fmt.Println("Warning: Unbounded solution detected")
    return -1, -1
  }
  
  return r, s
}

func (t *Tableau) PivotForFeasibility() (int, int) {
    n := len(t.Table[0]) // Number of columns
    m := len(t.Table)    // Number of rows
    
    // Find row with negative RHS
    r := -1
    for i := 0; i < m-1; i++ { // Skip objective row
        if t.Table[i][n-1].N < 0 && !t.dirtY[i] {
            r = i
            break
        }
    }
    
    if r == -1 {
        // No negative RHS found
        return -1, -1
    }
    
    // Find column with negative coefficient in that row
    s := -1
    mostNegative := fr.Fraction{N: 0, D: 1}
    
    for j := 0; j < n-1; j++ { // Skip constant column
        if t.Table[r][j].N < 0 && !t.dirtX[j] {
            // Choose the most negative coefficient
            if s == -1 || t.Table[r][j].N < mostNegative.N {
                s = j
                mostNegative = t.Table[r][j]
            }
        }
    }
    
    return r, s
}

func (t *Tableau) MakeFeasible() bool {
    fmt.Println("\nAttempting to make tableau feasible...")
    
    iteration := 1
    for !t.IsFeasible() {
        fmt.Printf("\n--- Feasibility Iteration %d ---\n", iteration)
        
        r, s := t.PivotForFeasibility()
        if !IsPivotValid(r, s) {
            fmt.Println("Failed to find appropriate pivot for feasibility. Problem may be infeasible.")
            return false
        }
        
        fmt.Printf("Pivoting on element at row %d, column %d (intersection of %s and %s)\n", 
                  r, s, t.RowNames[r], t.ColNames[s])
        *t = t.Transform(r, s)
        
        Print(t)
        iteration++
        
        // Safety check to prevent infinite loops
        if iteration > 20 {
            fmt.Println("Warning: Maximum feasibility iterations reached. Process stopped.")
            return false
        }
    }
    
    fmt.Println("Tableau is now feasible.")
    return true
}

func (t Tableau) Transform(r, s int) Tableau {
  b := t.Copy()
  b.dirtY[r] = true
  b.dirtX[s] = true

  pivotElement := t.Table[r][s]
  
  b.Table[r][s] = fr.Div(fr.Fraction{N: 1, D: 1}, pivotElement)
  
  for j := 0; j < len(t.Table[0]); j++ {
    if j != s {
      b.Table[r][j] = fr.Div(t.Table[r][j], pivotElement)
    }
  }
  
  for i := 0; i < len(t.Table); i++ {
    if i != r {
      for j := 0; j < len(t.Table[0]); j++ {
        if j != s {
          b.Table[i][j] = fillElement(t, i, j, r, s)
        }
      }
      b.Table[i][s] = fr.Neg(fr.Div(t.Table[i][s], pivotElement))
    }
  }

  b.RowNames[r], b.ColNames[s] = b.ColNames[s], b.RowNames[r]
  
  return b
}

func fillElement(a Tableau, i, j, r, s int) fr.Fraction {
  // Element transformation formula: (a_ij * a_rs - a_is * a_rj) / -a_rs
  pivotElement := a.Table[r][s]
  term1 := fr.Mul(a.Table[i][j], pivotElement)
  term2 := fr.Mul(a.Table[i][s], a.Table[r][j])
  result := fr.Div(fr.Sub(term1, term2), pivotElement)
  return result
}

func Print(a *Tableau) {
  fmt.Println("Current Tableau:")
  
  fmt.Printf("%-10s", "")
  for j := 0; j < len(a.Table[0]); j++ {
    if j < len(a.Table[0]) - 1 {
      fmt.Printf("%-10s", "-" + a.ColNames[j])
    } else {
      fmt.Printf("%-10s", a.ColNames[j])
    }
  }
  fmt.Println()
  
  for i := 0; i < len(a.Table); i++ {
    fmt.Printf("%-10s", a.RowNames[i])
    for j := 0; j < len(a.Table[i]); j++ {
      fr.Print(&a.Table[i][j], 10)
    }
    fmt.Println()
  }
}

func (a *Tableau) GetSolution() map[string]fr.Fraction {
  solution := make(map[string]fr.Fraction)
  m := len(a.Table)    // Number of rows
  n := len(a.Table[0]) // Number of columns
  
  solution["objective"] = a.Table[m-1][n-1]
  
  for i := 0; i < m-1; i++ {
    for j := 0; j < n-1; j++ {
      varName := a.RowNames[i]
      solution[varName] = a.Table[i][n-1]
    }
  }
  
  for j := 0; j < n-1; j++ {
    varName := a.ColNames[j]
    if _, exists := solution[varName]; !exists {
      solution[varName] = fr.Fraction{N: 0, D: 1}
    }
  }
  
  return solution
}

// Check if all RHS values are non-negative (feasible solution)
func (a *Tableau) IsFeasible() bool {
  n := len(a.Table[0])
  
  for i := 0; i < len(a.Table)-1; i++ {
    if a.Table[i][n-1].N < 0 {
      return false
    }
  }
  
  return true
}

// Check if optimal solution is reached
func (a *Tableau) IsOptimal() bool {
  n := len(a.Table[0]) // Number of columns
  m := len(a.Table)    // Number of rows
  
  for j := 0; j < n-1; j++ {
    if (a.IsMaximization && a.Table[m-1][j].N < 0) || 
       (!a.IsMaximization && a.Table[m-1][j].N > 0) {
      return false
    }
  }
  
  return true
}

func (a *Tableau) SetMaximization(isMax bool) {
  a.IsMaximization = isMax
}
