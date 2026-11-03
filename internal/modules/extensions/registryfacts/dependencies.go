package registryfacts

import "sort"

// Tarjan's algorithm returns maximal components, excluding vertices merely
// blocked by a cycle. Both traversal and emitted component members are stable.
func DependencyComponents(vertices []string, graph map[string][]string) [][]string {
	indices, low := map[string]int{}, map[string]int{}
	onStack := map[string]bool{}
	stack := []string{}
	components := [][]string{}
	index := 0
	var visit func(string)
	visit = func(id string) {
		index++
		indices[id], low[id] = index, index
		stack = append(stack, id)
		onStack[id] = true
		for _, target := range graph[id] {
			if indices[target] == 0 {
				visit(target)
				low[id] = min(low[id], low[target])
			} else if onStack[target] {
				low[id] = min(low[id], indices[target])
			}
		}
		if low[id] != indices[id] {
			return
		}
		component := []string{}
		for {
			top := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			onStack[top] = false
			component = append(component, top)
			if top == id {
				break
			}
		}
		sort.Strings(component)
		components = append(components, component)
	}
	for _, id := range vertices {
		if indices[id] == 0 {
			visit(id)
		}
	}
	return components
}
