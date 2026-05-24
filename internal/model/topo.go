package model

// topoSort performs topological sort on tables using Kahn's algorithm.
// It uses FK references to build the dependency graph: if table A has an FK
// referencing table B, then B must come before A.
// Returns sorted tables and any cycle groups (sets of mutually-referencing tables).
func topoSort(tables []Table) (sorted []Table, cycles [][]string) {
	tableByName := make(map[string]*Table, len(tables))
	for i := range tables {
		tableByName[tables[i].Name] = &tables[i]
	}

	// Build adjacency: inDegree counts how many FKs point into this table from others.
	// edges[A] = [B] means A depends on B (A has FK referencing B), so B must come first.
	inDegree := make(map[string]int, len(tables))
	dependsOn := make(map[string][]string, len(tables))

	for _, t := range tables {
		if _, ok := inDegree[t.Name]; !ok {
			inDegree[t.Name] = 0
		}
		for _, fk := range t.FKs {
			// Self-references don't create ordering constraints.
			if fk.RefTable == t.Name {
				continue
			}
			dependsOn[t.Name] = append(dependsOn[t.Name], fk.RefTable)
			inDegree[t.Name]++
		}
	}

	// Kahn's algorithm: start with zero-in-degree nodes.
	var queue []string
	for _, t := range tables {
		if inDegree[t.Name] == 0 {
			queue = append(queue, t.Name)
		}
	}

	visited := make(map[string]bool, len(tables))
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		visited[name] = true
		if t, ok := tableByName[name]; ok {
			sorted = append(sorted, *t)
		}

		// For each table that depends on this one, decrement its in-degree.
		for _, t := range tables {
			if visited[t.Name] {
				continue
			}
			for _, dep := range dependsOn[t.Name] {
				if dep == name {
					inDegree[t.Name]--
				}
			}
			if inDegree[t.Name] == 0 && !visited[t.Name] {
				// Check it's not already in the queue.
				if !inQueue(queue, t.Name) {
					queue = append(queue, t.Name)
				}
			}
		}
	}

	// Remaining nodes are in cycles.
	if len(sorted) < len(tables) {
		var cycleGroup []string
		for _, t := range tables {
			if !visited[t.Name] {
				cycleGroup = append(cycleGroup, t.Name)
				sorted = append(sorted, *tableByName[t.Name])
			}
		}
		if len(cycleGroup) > 0 {
			cycles = append(cycles, cycleGroup)
		}
	}

	return sorted, cycles
}

func inQueue(queue []string, name string) bool {
	for _, q := range queue {
		if q == name {
			return true
		}
	}
	return false
}
