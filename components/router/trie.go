package router

import (
	"strings"

	"github.com/go-zoox/core-utils/safe"
)

// Node ...
type Node struct {
	Path     string
	part     string
	children []*Node
	isWild   bool
	//
	paramsKeyIndexMap *safe.Map
}

// MatchChild ...
func (n *Node) MatchChild(part string) *Node {
	for _, child := range n.children {
		for child.part == part || child.isWild {
			return child
		}
	}

	return nil
}

func (n *Node) matchChildren(part string) []*Node {
	nodes := []*Node{}

	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}

	return nodes
}

// Insert ...
func (n *Node) Insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		n.Path = pattern
		return
	}

	paramsKeyIndexMap := safe.NewMap()
	for i, part := range parts {
		if part[0] == ':' {
			// pattern: /user/:name
			paramsKeyIndexMap.Set(part[1:], i)
		} else if part[0] == '{' && part[len(part)-1] == '}' {
			// pattern: /user/{name}
			paramsKeyIndexMap.Set(part[1:len(part)-1], i)
		} else if part[0] == '*' && len(part) > 1 {
			// pattern: /file/*
			paramsKeyIndexMap.Set(part[1:], i)
			break
		}
	}

	part := parts[height]
	child := n.MatchChild(part)
	if child == nil {
		isWild := false
		if part[0] == ':' {
			isWild = true
		} else if part[0] == '{' && part[len(part)-1] == '}' {
			isWild = true
		} else if part[0] == '*' {
			isWild = true
		}

		child = &Node{
			part:              part,
			isWild:            isWild,
			paramsKeyIndexMap: paramsKeyIndexMap,
		}
		n.children = append(n.children, child)
	}

	child.Insert(pattern, parts, height+1)
}

// Search ...
func (n *Node) Search(parts []string, height int) *Node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.Path == "" {
			return nil
		}

		return n
	}

	part := parts[height]
	children := n.matchChildren(part)

	for _, child := range children {
		result := child.Search(parts, height+1)
		if result != nil {
			return result
		}
	}

	return nil
}

func (n *Node) IsWild() bool {
	return n.isWild
}

func (n *Node) ParamsMap() *safe.Map {
	return n.paramsKeyIndexMap
}
