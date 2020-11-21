package ast

import (
	"bytes"
	"fmt"
	"monkey/token"
	"strings"
)

type Identifier struct {
	Token token.Token // INT
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

type IntegerLiteral struct {
	Token token.Token // INT
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

type PrefixExpression struct {
	Token    token.Token // e.g. "!", "-"
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("(%s%s)", pe.Operator, pe.Right.String()))
	return out.String()
}

type InfixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
	Left     Expression
}

func (oe *InfixExpression) expressionNode()      {}
func (oe *InfixExpression) TokenLiteral() string { return oe.Token.Literal }
func (oe *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("(%s %s %s)", oe.Left.String(), oe.Operator, oe.Right.String()))
	return out.String()
}

type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }

type IfExpression struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer
	out.WriteString(fmt.Sprintf("if%s %s", ie.Condition.String(), ie.Consequence.String()))

	if ie.Alternative != nil {
		out.WriteString(fmt.Sprintf("else %s", ie.Alternative.String()))
	}
	return out.String()
}

type FuncLiteral struct {
	Token      token.Token // fn token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FuncLiteral) expressionNode()      {}
func (fl *FuncLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FuncLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(fmt.Sprintf("%s(%s) %s", fl.Token.Literal, strings.Join(params, ", "), fl.Body.String()))
	return out.String()
}

type CallExpression struct {
	Token     token.Token // "(" is the infix operator
	Function  Expression  // identifier or func literal
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, p := range ce.Arguments {
		args = append(args, p.String())
	}
	out.WriteString(fmt.Sprintf("%s(%s)", ce.Function.String(), strings.Join(args, ", ")))
	return out.String()
}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (s *StringLiteral) expressionNode()      {}
func (s *StringLiteral) TokenLiteral() string { return s.Token.Literal }
func (s *StringLiteral) String() string       { return s.Value }
