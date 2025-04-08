package fraction

import (
	"fmt"
	"strconv"
	"strings"
)

type Fraction struct {
  N int
  D int
}

func (n Fraction) gcd() int {
  a := max(n.D, n.N)
  b := min(n.D, n.N)
  if b == 0 || a == 0 { return 1 }
  f := 0
  for {
    if f = a % b; f == 0 { return b }
    a = b
    b = f
  }
}

func (n *Fraction) Simplify() {
  f := n.gcd();
  n.D /= f
  n.N /= f

  if n.D < 0 {
    n.D *= -1
    n.N *= -1
  }
}

func (n *Fraction) reverse() {
  n.N = n.N ^ n.D
  n.D = n.N ^ n.D
  n.N = n.N ^ n.D
}

func Read(n *Fraction) {
  var text string
  fmt.Scan(&text)
  text = strings.TrimSpace(text)
  parts := strings.Split(text, "/")
  if len(parts) > 0 {
    n.N, _ = strconv.Atoi(parts[0])
    n.D = 1
  }
  if len(parts) > 1 {
    n.D, _ = strconv.Atoi(parts[1])
  }

  n.Simplify()
}

func Print(n *Fraction, p uint) {
  var buffer string
  switch {
  case n.N < 0 && n.D == 1:
    buffer = fmt.Sprintf("%d", n.N)
  case n.N < 0:
    buffer = fmt.Sprintf("%d/%d", n.N, n.D)
  case n.D == 1:
    buffer = fmt.Sprintf(" %d", n.N)
  default:
    buffer = fmt.Sprintf(" %d/%d", n.N, n.D)
  }

  fmt.Printf("%-*s", p, buffer)
}

func Add(a, b Fraction) (res Fraction) {
  res.N = a.N * b.D + a.D * b.N
  res.D = a.D * b.D
  res.Simplify()
  return
}

func Sub(a, b Fraction) (res Fraction) {
  res.N = a.N * b.D - b.N * a.D
  res.D = a.D * b.D
  res.Simplify()
  return
}

func Mul(a, b Fraction) (res Fraction) {
  res.N = a.N * b.N
  res.D = a.D * b.D
  res.Simplify()
  return
}

func Div(a, b Fraction) (res Fraction) {
  res.N = a.N * b.D
  res.D = a.D * b.N
  res.Simplify()
  return
}

func Neg(a Fraction) (res Fraction) {
  res = Mul(a, Fraction{-1, 1})
  return
}
