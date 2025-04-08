// Package parser handles parsing of linear programming problems
package parser

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	fr "simplex/fraction"
	tb "simplex/tableau"
)

// Term represents a single term in an equation like 3x1 or -2x2
type Term struct {
	Coefficient fr.Fraction
	Variable    string
}

// Equation represents a linear equation or inequality
type Equation struct {
	LHS      []Term
	RHS      fr.Fraction
	Relation string // "<=", ">=", "="
}

// Problem represents a complete linear programming problem
type Problem struct {
	ObjectiveFunction Equation
	Constraints       []Equation
	IsMaximization    bool
	Variables         map[string]bool // Set of all variables
}

// ParseProblem parses a complete linear programming problem
func ParseProblem(objectiveStr string, constraintStrs []string, isMax bool) (*Problem, error) {
	problem := &Problem{
		IsMaximization: isMax,
		Constraints:    make([]Equation, 0, len(constraintStrs)),
		Variables:      make(map[string]bool),
	}

	// Parse objective function (format: "3x1 + 2x2 + ... + 5xn")
	obj, err := parseEquation(objectiveStr, "=")
	if err != nil {
		return nil, fmt.Errorf("error parsing objective function: %w", err)
	}
	problem.ObjectiveFunction = obj

	// Add variables from objective function
	for _, term := range obj.LHS {
		if term.Variable != "" {
			problem.Variables[term.Variable] = true
		}
	}

	// Parse constraints
	for i, constraintStr := range constraintStrs {
		// First find the relation
		relation := ""
		for _, rel := range []string{"<=", ">=", "="} {
			if strings.Contains(constraintStr, rel) {
				relation = rel
				break
			}
		}
		if relation == "" {
			return nil, fmt.Errorf("constraint %d does not contain a valid relation (<=, >=, =)", i+1)
		}

		constraint, err := parseEquation(constraintStr, relation)
		if err != nil {
			return nil, fmt.Errorf("error parsing constraint %d: %w", i+1, err)
		}
		problem.Constraints = append(problem.Constraints, constraint)

		// Add variables from constraints
		for _, term := range constraint.LHS {
			if term.Variable != "" {
				problem.Variables[term.Variable] = true
			}
		}
	}

	return problem, nil
}

// ParseEquation parses a single equation or inequality
func parseEquation(eqStr, relation string) (Equation, error) {
	parts := strings.SplitN(eqStr, relation, 2)
	if len(parts) != 2 {
		// If no relation is found (for objective function), assume RHS is 0
		if relation == "=" && !strings.Contains(eqStr, "=") {
			parts = []string{eqStr, "0"}
		} else {
			return Equation{}, fmt.Errorf("invalid equation format: %s", eqStr)
		}
	}

	lhsStr := strings.TrimSpace(parts[0])
	rhsStr := strings.TrimSpace(parts[1])

	// Parse RHS
	rhsFraction, err := parseFraction(rhsStr)
	if err != nil {
		return Equation{}, fmt.Errorf("invalid right-hand side: %s", rhsStr)
	}

	// Parse LHS
	terms, err := parseTerms(lhsStr)
	if err != nil {
		return Equation{}, fmt.Errorf("error parsing left-hand side: %w", err)
	}

	return Equation{
		LHS:      terms,
		RHS:      rhsFraction,
		Relation: relation,
	}, nil
}

// ParseTerms parses the terms in the left-hand side of an equation
func parseTerms(lhsStr string) ([]Term, error) {
	// Add + between terms if there's no operator
	re := regexp.MustCompile(`([^\+\-\s])(\s+)(-?\d)`)
	lhsStr = re.ReplaceAllString(lhsStr, "$1 + $3")

	// Add + at the beginning if the expression starts with a variable
	if !strings.HasPrefix(strings.TrimSpace(lhsStr), "+") && !strings.HasPrefix(strings.TrimSpace(lhsStr), "-") {
		lhsStr = "+ " + lhsStr
	}

	// Normalize spaces around operators
	lhsStr = regexp.MustCompile(`\s*([\+\-])\s*`).ReplaceAllString(lhsStr, " $1 ")
	lhsStr = strings.TrimSpace(lhsStr)

	// Split by + and - operators
	components := strings.FieldsFunc(lhsStr, func(r rune) bool {
		return r == '+' || r == '-'
	})

	// Find operators (+ or -)
	operators := make([]string, 0, len(components))
	opRegex := regexp.MustCompile(`[\+\-]`)
	ops := opRegex.FindAllString(lhsStr, -1)
	operators = append(operators, ops...)

	if len(components) != len(operators) {
		return nil, errors.New("parsing error: mismatch between terms and operators")
	}

	terms := make([]Term, 0, len(components))

	// Process each term with its operator
	for i, component := range components {
		component = strings.TrimSpace(component)
		if component == "" {
			continue
		}

		// Get the sign from the operator
		sign := 1
		if operators[i] == "-" {
			sign = -1
		}

		// Split coefficient and variable
		var coef fr.Fraction
		var variable string

		// Check if component is just a number
		if _, err := strconv.Atoi(component); err == nil {
			coef = fr.Fraction{N: sign * mustAtoi(component), D: 1}
			variable = "" // No variable associated
		} else {
			// Match pattern like "2x1", "x2", etc.
			re = regexp.MustCompile(`^(\d*)(x\d+|[a-zA-Z]\d*)$`)
			matches := re.FindStringSubmatch(component)

			if len(matches) == 3 {
				coefStr := matches[1]
				variable = matches[2]

				if coefStr == "" {
					// If no coefficient is specified, it's 1
					coef = fr.Fraction{N: sign * 1, D: 1}
				} else {
					coef = fr.Fraction{N: sign * mustAtoi(coefStr), D: 1}
				}
			} else {
				// Check for fractions like "1/2x1"
				re = regexp.MustCompile(`^(\d+)/(\d+)(x\d+|[a-zA-Z]\d*)$`)
				matches = re.FindStringSubmatch(component)

				if len(matches) == 4 {
					num := sign * mustAtoi(matches[1])
					den := mustAtoi(matches[2])
					variable = matches[3]
					coef = fr.Fraction{N: num, D: den}
					coef.Simplify()
				} else {
					return nil, fmt.Errorf("invalid term format: %s", component)
				}
			}
		}

		terms = append(terms, Term{Coefficient: coef, Variable: variable})
	}

	return terms, nil
}

// ParseFraction parses a string into a Fraction
func parseFraction(s string) (fr.Fraction, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return fr.Fraction{N: 0, D: 1}, nil
	}

	// Check if it's a simple integer
	if i, err := strconv.Atoi(s); err == nil {
		return fr.Fraction{N: i, D: 1}, nil
	}

	// Check if it's a fraction like "1/2"
	parts := strings.Split(s, "/")
	if len(parts) == 2 {
		num, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		den, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err1 == nil && err2 == nil && den != 0 {
			frac := fr.Fraction{N: num, D: den}
			frac.Simplify()
			return frac, nil
		}
	}

	return fr.Fraction{}, fmt.Errorf("invalid number format: %s", s)
}

// ConvertToTableau converts a Problem to a Tableau in standard form for simplex method
func ConvertToTableau(p *Problem) tb.Tableau {
	// Extract all decision variables from the problem
	decisionVars := make([]string, 0, len(p.Variables))
	for v := range p.Variables {
		decisionVars = append(decisionVars, v)
	}
	// Sort variables for consistent ordering
	sort.Strings(decisionVars)

	// Create a tableau with the appropriate dimensions
	// Rows: one for each constraint plus objective function
	// Columns: one for each decision variable plus RHS
	numRows := len(p.Constraints) + 1
	numCols := len(decisionVars) + 1 // +1 for RHS
	var t tb.Tableau
	t.Init(numRows, numCols)
	t.SetMaximization(p.IsMaximization)

	// Set up column names (decision variables)
	for i, v := range decisionVars {
		// In standard simplex tableau, we use negative of variables
		t.ColNames[i] = "-" + v
	}
	t.ColNames[numCols-1] = "const" // Last column is constants

	// Set up row names (slack variables)
	for i := 0; i < len(p.Constraints); i++ {
		t.RowNames[i] = fmt.Sprintf("s%d", i+1)
	}
	t.RowNames[numRows-1] = "F" // Last row is objective function

	// Fill in constraint rows
	for i, constraint := range p.Constraints {
		// Initialize RHS with constraint's RHS
		t.Table[i][numCols-1] = constraint.RHS

		// Add coefficients for decision variables with proper signs
		for _, term := range constraint.LHS {
			if term.Variable != "" {
				// Find corresponding column
				for j, v := range decisionVars {
					if v == term.Variable {
						// For standard form, we move all variables to RHS with negated coefficients
						// But in tableau, we keep the original sign for computational purposes
						t.Table[i][j] = term.Coefficient
						break
					}
				}
			} else {
				// Constant term is handled by adjusting RHS
				t.Table[i][numCols-1] = fr.Sub(t.Table[i][numCols-1], term.Coefficient)
			}
		}

		// Handle inequality relations
		if constraint.Relation == ">=" {
			// For >= constraint, negate entire row to make it <= form
			for j := 0; j < numCols; j++ {
				t.Table[i][j] = fr.Neg(t.Table[i][j])
			}
		}
		// For = constraints, we keep them as is
	}

	// Fill in objective function row
	objRow := numRows - 1
	for _, term := range p.ObjectiveFunction.LHS {
		if term.Variable != "" {
			// Find corresponding column
			for j, v := range decisionVars {
				if v == term.Variable {
					if p.IsMaximization {
						// For maximization, we put negative coefficients in objective row
						t.Table[objRow][j] = fr.Neg(term.Coefficient)
					} else {
						// For minimization, coefficient signs remain unchanged
						t.Table[objRow][j] = term.Coefficient
					}
					break
				}
			}
		} else {
			// Constant term goes to RHS
			if p.IsMaximization {
				t.Table[objRow][numCols-1] = fr.Add(t.Table[objRow][numCols-1], term.Coefficient)
			} else {
				t.Table[objRow][numCols-1] = fr.Sub(t.Table[objRow][numCols-1], term.Coefficient)
			}
		}
	}

	return t
}

// Helper function to convert string to int, panics on error
func mustAtoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Sprintf("mustAtoi: cannot convert %s to int", s))
	}
	return i
}
