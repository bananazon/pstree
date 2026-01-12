package pstree

// OptimizedBuildTree is an optimized version of BuildTree that uses the PidToIndexMap
// to avoid linear searches through the process list.
func (processTree *ProcessTree) OptimizedBuildTree() {
	processTree.Logger.Debug("Entering processTree.OptimizedBuildTree()")

	// Initialize all nodes with -1 for Child, Parent, and Sister fields
	for i := range processTree.Nodes {
		processTree.Nodes[i].Child = -1
		processTree.Nodes[i].Parent = -1
		processTree.Nodes[i].Sister = -1
	}

	// Build the tree using the PidToIndexMap for O(1) lookups
	for pidIndex := range processTree.Nodes {
		ppid := processTree.Nodes[pidIndex].PPID
		
		// Look up parent index directly from the map
		ppidIndex, exists := processTree.PidToIndexMap[ppid]
		
		// Skip if parent doesn't exist or is the process itself
		if !exists || ppidIndex == pidIndex {
			continue
		}
		
		// Set parent relationship
		processTree.Nodes[pidIndex].Parent = ppidIndex
		
		// Add as child
		if processTree.Nodes[ppidIndex].Child == -1 {
			// First child
			processTree.Nodes[ppidIndex].Child = pidIndex
		} else {
			// Find the last sibling
			sisterIndex := processTree.Nodes[ppidIndex].Child
			for processTree.Nodes[sisterIndex].Sister != -1 {
				sisterIndex = processTree.Nodes[sisterIndex].Sister
			}
			// Add as sister to the last child
			processTree.Nodes[sisterIndex].Sister = pidIndex
		}
	}
}

// ReplaceWithOptimizedBuildTree replaces the original BuildTree method with the optimized version
func ReplaceWithOptimizedBuildTree() {
	// This is a placeholder function that could be used to monkey patch
	// the original BuildTree with the optimized version if needed
}
