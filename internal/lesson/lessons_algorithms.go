package lesson

// algorithmsLessons covers the interview classics: two pointers, sliding window, binary search, backtracking, DP, BFS/DFS, heaps, linked lists, and LRU cache.
var algorithmsLessons = []Lesson{
	{
		ID:       "algo-two-pointers",
		Category: "Interview Algorithms",
		Title:    "Two Pointers",
		Description: `
<p>The two-pointers technique walks a sequence with two indices,
typically converging toward each other or moving at different speeds.
Classic problems: <b>reverse a string</b>, <b>valid palindrome</b>,
<b>two-sum on sorted array</b>, <b>3-sum</b>, <b>container with most
water</b>.</p>

<p>Time complexity drops from O(n²) (brute force) to O(n). Memory is O(1).</p>
`,
		Code: `package main

import "fmt"

// twoSumSorted: given a SORTED slice and target, return (i,j,found).
// O(n) time, O(1) space.
func twoSumSorted(nums []int, target int) (int, int, bool) {
	lo, hi := 0, len(nums)-1
	for lo < hi {
		sum := nums[lo] + nums[hi]
		switch {
		case sum == target:
			return lo, hi, true
		case sum < target:
			lo++
		default:
			hi--
		}
	}
	return 0, 0, false
}

// isPalindrome: case-insensitive, ignore non-alphanumeric.
func isPalindrome(s string) bool {
	b := []byte(s)
	lo, hi := 0, len(b)-1
	for lo < hi {
		for lo < hi && !isAlnum(b[lo]) {
			lo++
		}
		for lo < hi && !isAlnum(b[hi]) {
			hi--
		}
		if lower(b[lo]) != lower(b[hi]) {
			return false
		}
		lo++
		hi--
	}
	return true
}

func isAlnum(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func lower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + 32
	}
	return b
}

func main() {
	i, j, ok := twoSumSorted([]int{1, 3, 4, 5, 7, 11}, 9)
	fmt.Println("twoSumSorted ->", i, j, ok) // 2 3 true  (4 + 5 = 9)

	fmt.Println(isPalindrome("A man, a plan, a canal: Panama")) // true
	fmt.Println(isPalindrome("race a car"))                     // false
}
`,
		Notes: []string{
			"Sorted input is the signal to reach for two pointers. If unsorted, consider a hash map first.",
			"Sliding window is a sibling pattern — two pointers moving the SAME direction.",
			"3-sum: sort, fix the first element, two-pointer on the rest. O(n²) total.",
			"Edge cases to always test: empty slice, all duplicates, target unreachable.",
		},
	},
	{
		ID:       "algo-sliding-window",
		Category: "Interview Algorithms",
		Title:    "Sliding Window",
		Description: `
<p>A sliding window maintains a contiguous sub-array/sub-string and
moves the boundaries to satisfy a constraint. Two variants:</p>

<ul>
  <li><b>Fixed-size window</b> — both endpoints advance together. Use
  for "max sum of size-k sub-array."</li>
  <li><b>Variable-size window</b> — expand the right endpoint, contract
  the left when a constraint is violated. Use for "longest substring
  without repeating characters."</li>
</ul>

<p>O(n) time despite looking nested, because each pointer advances at
most n times total — classic amortized analysis.</p>
`,
		Code: `package main

import "fmt"

// Longest substring without repeating chars. O(n).
func longestUnique(s string) int {
	last := make(map[byte]int) // char -> last index seen
	best, left := 0, 0
	for right := 0; right < len(s); right++ {
		if idx, seen := last[s[right]]; seen && idx >= left {
			left = idx + 1 // shrink: jump past the previous occurrence
		}
		last[s[right]] = right
		if right-left+1 > best {
			best = right - left + 1
		}
	}
	return best
}

// Max sum of any contiguous sub-array of size k. O(n).
func maxSumK(nums []int, k int) int {
	if len(nums) < k {
		return 0
	}
	sum := 0
	for _, v := range nums[:k] {
		sum += v
	}
	best := sum
	for i := k; i < len(nums); i++ {
		sum += nums[i] - nums[i-k]
		if sum > best {
			best = sum
		}
	}
	return best
}

func main() {
	fmt.Println(longestUnique("abcabcbb")) // 3 ("abc")
	fmt.Println(longestUnique("bbbbb"))    // 1
	fmt.Println(longestUnique("pwwkew"))   // 3 ("wke")

	fmt.Println(maxSumK([]int{1, 4, 2, 10, 23, 3, 1, 0, 20}, 4)) // 39
}
`,
		Notes: []string{
			"Signal: \"longest/shortest/max/min contiguous sub-array satisfying X.\"",
			"Map + window handles \"no repeats / at most K distinct chars\" variants.",
			"For numeric problems, prefix sums are the alternative — pick whichever is cleaner.",
			"Track `best` INSIDE the loop, not after — easy to forget the final window.",
		},
	},
	{
		ID:       "algo-binary-search",
		Category: "Interview Algorithms",
		Title:    "Binary Search & Variants",
		Description: `
<p>Vanilla binary search is easy. Senior interviews ask the VARIANTS:</p>

<ul>
  <li><b>Leftmost</b> insertion (first index where x ≥ target).</li>
  <li><b>Rightmost</b> insertion (one past last index where x ≤ target).</li>
  <li><b>Binary search on the ANSWER</b> — candidate space is monotonic
  but not an explicit array (e.g., "min capacity to ship in K days").</li>
  <li><b>Rotated sorted array</b> — search a sorted-then-rotated slice
  in O(log n).</li>
</ul>

<p>Subtle point interviewers love: <code>mid := lo + (hi-lo)/2</code>,
not <code>(lo+hi)/2</code>, to avoid integer overflow.</p>

<p>Standard library: <code>sort.Search</code> and
<code>slices.BinarySearch</code> (1.21+).</p>
`,
		Code: `package main

import (
	"fmt"
	"sort"
)

// Leftmost: smallest i such that nums[i] >= target.
func lowerBound(nums []int, target int) int {
	lo, hi := 0, len(nums)
	for lo < hi {
		mid := lo + (hi-lo)/2
		if nums[mid] < target {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}

// Search rotated sorted array. O(log n).
func searchRotated(nums []int, target int) int {
	lo, hi := 0, len(nums)-1
	for lo <= hi {
		mid := lo + (hi-lo)/2
		if nums[mid] == target {
			return mid
		}
		if nums[lo] <= nums[mid] { // left half sorted
			if nums[lo] <= target && target < nums[mid] {
				hi = mid - 1
			} else {
				lo = mid + 1
			}
		} else { // right half sorted
			if nums[mid] < target && target <= nums[hi] {
				lo = mid + 1
			} else {
				hi = mid - 1
			}
		}
	}
	return -1
}

func main() {
	nums := []int{1, 3, 3, 5, 7, 9}
	fmt.Println("lowerBound(3):", lowerBound(nums, 3)) // 1
	fmt.Println("lowerBound(4):", lowerBound(nums, 4)) // 3

	fmt.Println("sort.SearchInts(3):", sort.SearchInts(nums, 3)) // 1

	fmt.Println("rotated search:", searchRotated([]int{4, 5, 6, 7, 0, 1, 2}, 0)) // 4
}
`,
		Notes: []string{
			"Use [lo, hi) half-open intervals — fewer off-by-one bugs than [lo, hi].",
			"sort.Search takes a predicate: sort.Search(n, func(i int) bool { return cond(i) }).",
			"Binary search on the answer: predicate must be monotonic — false-false-...-true-true.",
			"Rotated array: identify which half is sorted by comparing nums[lo] to nums[mid].",
		},
	},
	{
		ID:       "algo-backtracking",
		Category: "Interview Algorithms",
		Title:    "Backtracking (Subsets, Permutations)",
		Description: `
<p>Backtracking = DFS over a decision tree, with state mutated in-place
and "undone" as we unwind. Templates:</p>

<ul>
  <li><b>Subsets / Combinations:</b> include or skip each element.</li>
  <li><b>Permutations:</b> at each step pick any unused element.</li>
  <li><b>N-Queens / Word Search:</b> place a piece, recurse, remove it.</li>
</ul>

<p>Most backtracking solutions are 15 lines of Go and exponential time
worst-case — that's expected.</p>
`,
		Code: `package main

import "fmt"

// All subsets of nums. 2^n.
func subsets(nums []int) [][]int {
	var out [][]int
	var path []int
	var dfs func(int)
	dfs = func(i int) {
		if i == len(nums) {
			cp := append([]int(nil), path...) // copy before storing
			out = append(out, cp)
			return
		}
		dfs(i + 1) // skip
		path = append(path, nums[i])
		dfs(i + 1) // take
		path = path[:len(path)-1]
	}
	dfs(0)
	return out
}

// All permutations of nums. n!.
func permute(nums []int) [][]int {
	var out [][]int
	used := make([]bool, len(nums))
	var path []int
	var dfs func()
	dfs = func() {
		if len(path) == len(nums) {
			cp := append([]int(nil), path...)
			out = append(out, cp)
			return
		}
		for i, v := range nums {
			if used[i] {
				continue
			}
			used[i] = true
			path = append(path, v)
			dfs()
			path = path[:len(path)-1]
			used[i] = false
		}
	}
	dfs()
	return out
}

func main() {
	fmt.Println("subsets:", subsets([]int{1, 2, 3}))
	fmt.Println("permutations:", permute([]int{1, 2, 3}))
}
`,
		Notes: []string{
			"ALWAYS copy `path` before appending to results — it's mutated in place.",
			"For \"unique\" variants, sort the input first and skip duplicates at the SAME tree depth.",
			"Pruning is the senior move: `if currentSum > target { return }` early-exits dead branches.",
			"Know the runtimes cold: subsets 2^n, permutations n!, combinations C(n,k).",
		},
	},
	{
		ID:       "algo-dp",
		Category: "Interview Algorithms",
		Title:    "Dynamic Programming Essentials",
		Description: `
<p>DP recipe:</p>

<ol>
  <li><b>State</b> — what does dp[i] (or dp[i][j]) mean? Write it in English.</li>
  <li><b>Transition</b> — how does dp[i] depend on smaller states?</li>
  <li><b>Base case</b> — what's dp[0]?</li>
  <li><b>Order</b> — iterate so dependencies are computed first.</li>
</ol>

<p>Three patterns to know:</p>
<ul>
  <li><b>1D DP</b> — climbing stairs, house robber, coin change.</li>
  <li><b>2D DP</b> — longest common subsequence, edit distance, unique paths.</li>
  <li><b>Knapsack</b> — pick items subject to a capacity.</li>
</ul>
`,
		Code: `package main

import "fmt"

// Min coins to make amount. -1 if impossible. Unbounded knapsack.
// dp[a] = fewest coins for amount a; dp[a] = min(dp[a-c]+1) over coins c.
func coinChange(coins []int, amount int) int {
	const inf = 1 << 30
	dp := make([]int, amount+1)
	for i := range dp {
		dp[i] = inf
	}
	dp[0] = 0
	for a := 1; a <= amount; a++ {
		for _, c := range coins {
			if a-c >= 0 && dp[a-c]+1 < dp[a] {
				dp[a] = dp[a-c] + 1
			}
		}
	}
	if dp[amount] == inf {
		return -1
	}
	return dp[amount]
}

// Longest common subsequence of two strings. 2D DP, O(m*n).
func lcs(a, b string) int {
	m, n := len(a), len(b)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] > dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}
	return dp[m][n]
}

func main() {
	fmt.Println("coinChange([1,2,5], 11):", coinChange([]int{1, 2, 5}, 11)) // 3
	fmt.Println("coinChange([2], 3):    ", coinChange([]int{2}, 3))         // -1
	fmt.Println("lcs ABCBDAB / BDCABA:  ", lcs("ABCBDAB", "BDCABA"))        // 4
}
`,
		Notes: []string{
			"Articulate the state in English BEFORE coding. \"dp[i] = fewest coins for amount i.\"",
			"Memoization (top-down) and tabulation (bottom-up) give equivalent results — pick what's easier to derive.",
			"Many 2D DPs space-optimize to two rows or even one. Worth mentioning in interviews.",
			"Can't derive the recurrence? Hand-trace a tiny example first — the pattern jumps out.",
		},
	},
	{
		ID:       "algo-bfs-dfs",
		Category: "Interview Algorithms",
		Title:    "Graph Traversal (BFS / DFS)",
		Description: `
<p>BFS for shortest paths in unweighted graphs; DFS for almost everything
else (cycles, components, topological sort, articulation points). Both
are O(V+E). Both should be 20-line muscle memory by interview day.</p>

<h3>Representation</h3>
<ul>
  <li><b>Adjacency list</b> — <code>map[int][]int</code> or <code>[][]int</code>. Default pick.</li>
  <li><b>Grid as implicit graph</b> — for 2D problems, neighbors are the 4 (or 8) directions.</li>
</ul>
`,
		Code: `package main

import "fmt"

type Graph struct {
	adj map[int][]int
}

func NewGraph() *Graph { return &Graph{adj: make(map[int][]int)} }

func (g *Graph) AddEdge(u, v int) {
	g.adj[u] = append(g.adj[u], v)
	g.adj[v] = append(g.adj[v], u) // undirected
}

// BFS: shortest path (in edges) from src to each reachable node.
func (g *Graph) BFS(src int) map[int]int {
	dist := map[int]int{src: 0}
	q := []int{src}
	for len(q) > 0 {
		u := q[0]
		q = q[1:]
		for _, v := range g.adj[u] {
			if _, seen := dist[v]; !seen {
				dist[v] = dist[u] + 1
				q = append(q, v)
			}
		}
	}
	return dist
}

// DFS (iterative) — collect the connected component containing src.
func (g *Graph) Component(src int) []int {
	var out []int
	seen := map[int]bool{}
	stack := []int{src}
	for len(stack) > 0 {
		u := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[u] {
			continue
		}
		seen[u] = true
		out = append(out, u)
		for _, v := range g.adj[u] {
			if !seen[v] {
				stack = append(stack, v)
			}
		}
	}
	return out
}

func main() {
	g := NewGraph()
	for _, e := range [][2]int{{1, 2}, {1, 3}, {2, 4}, {3, 4}, {4, 5}, {6, 7}} {
		g.AddEdge(e[0], e[1])
	}
	fmt.Println("BFS dists from 1:", g.BFS(1))
	fmt.Println("Component of 1: ", g.Component(1))
	fmt.Println("Component of 6: ", g.Component(6))
}
`,
		Notes: []string{
			"BFS uses a queue (slice + head pop); DFS uses a stack or recursion.",
			"For grids: precompute `dirs := [4][2]int{{1,0},{-1,0},{0,1},{0,-1}}` — clean loop body.",
			"Topological sort: DFS with postorder stack OR Kahn's algorithm (in-degree BFS).",
			"Visited set is mandatory — without it, cycles → infinite loop.",
		},
	},
	{
		ID:       "algo-heap",
		Category: "Interview Algorithms",
		Title:    "Heaps & Top-K Problems",
		Description: `
<p>"Top-K largest", "K-closest points", "median of a stream",
"merge K sorted lists" — all heap problems. Go ships
<code>container/heap</code>: implement five methods (Len, Less, Swap,
Push, Pop) and call <code>heap.Push</code> / <code>heap.Pop</code> /
<code>heap.Init</code>.</p>

<p>For top-K largest: use a MIN-heap of size K (counterintuitive but
correct — the heap holds your K candidates, root is the smallest of
them; replace if a new element is larger).</p>
`,
		Code: `package main

import (
	"container/heap"
	"fmt"
)

// IntMinHeap implements heap.Interface as a min-heap.
type IntMinHeap []int

func (h IntMinHeap) Len() int           { return len(h) }
func (h IntMinHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h IntMinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *IntMinHeap) Push(x any)        { *h = append(*h, x.(int)) }
func (h *IntMinHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// Top K largest. O(n log k) time, O(k) space.
func topK(nums []int, k int) []int {
	h := &IntMinHeap{}
	heap.Init(h)
	for _, n := range nums {
		if h.Len() < k {
			heap.Push(h, n)
		} else if n > (*h)[0] {
			heap.Pop(h)
			heap.Push(h, n)
		}
	}
	return *h
}

func main() {
	fmt.Println("top 3:", topK([]int{3, 2, 1, 5, 6, 4, 9, 0, 7, 8}, 3)) // {7 8 9} in some order
}
`,
		Notes: []string{
			"heap.Push / heap.Pop are PACKAGE functions; your type's methods are primitives for the package.",
			"Min-heap of size K → top-K LARGEST. Max-heap of size K → top-K SMALLEST. Yes, swapped.",
			"Streaming median: TWO heaps — max-heap of lower half, min-heap of upper, keep sizes balanced.",
			"heap.Init is O(n). Pushing n elements one by one is O(n log n).",
		},
	},
	{
		ID:       "algo-linked-list",
		Category: "Interview Algorithms",
		Title:    "Linked List Patterns",
		Description: `
<p>Three problems show up over and over: <b>reverse</b>, <b>detect
cycle</b> (Floyd's tortoise & hare), and <b>merge two sorted lists</b>.
Master these three and ~70% of linked-list questions are trivial
variants.</p>

<p>Pro tip: the <b>dummy head node</b> trick eliminates almost all edge
cases when building a list.</p>
`,
		Code: `package main

import "fmt"

type Node struct {
	Val  int
	Next *Node
}

func fromSlice(xs []int) *Node {
	dummy := &Node{}
	tail := dummy
	for _, x := range xs {
		tail.Next = &Node{Val: x}
		tail = tail.Next
	}
	return dummy.Next
}

func dump(head *Node) {
	for n := head; n != nil; n = n.Next {
		fmt.Printf("%d ", n.Val)
	}
	fmt.Println()
}

// Reverse in place. O(n) time, O(1) space.
func reverse(head *Node) *Node {
	var prev *Node
	for cur := head; cur != nil; {
		next := cur.Next
		cur.Next = prev
		prev = cur
		cur = next
	}
	return prev
}

// Floyd's tortoise & hare: cycle detection.
func hasCycle(head *Node) bool {
	slow, fast := head, head
	for fast != nil && fast.Next != nil {
		slow = slow.Next
		fast = fast.Next.Next
		if slow == fast {
			return true
		}
	}
	return false
}

// Merge two sorted lists.
func merge(a, b *Node) *Node {
	dummy := &Node{}
	tail := dummy
	for a != nil && b != nil {
		if a.Val <= b.Val {
			tail.Next = a
			a = a.Next
		} else {
			tail.Next = b
			b = b.Next
		}
		tail = tail.Next
	}
	if a != nil {
		tail.Next = a
	} else {
		tail.Next = b
	}
	return dummy.Next
}

func main() {
	l := fromSlice([]int{1, 2, 3, 4, 5})
	dump(reverse(l)) // 5 4 3 2 1

	a := fromSlice([]int{1, 3, 5})
	b := fromSlice([]int{2, 4, 6})
	dump(merge(a, b)) // 1 2 3 4 5 6

	fmt.Println("cycle in [1->2->3]?", hasCycle(fromSlice([]int{1, 2, 3})))
}
`,
		Notes: []string{
			"Dummy-head pattern eliminates the \"first node is special\" edge case.",
			"Floyd's cycle detection — slow takes 1 step, fast takes 2; they meet inside the cycle.",
			"To find the cycle START: after they meet, move one pointer back to head; advance both 1 step at a time.",
			"Reversing a list: ALWAYS save Next first; otherwise you lose the rest of the list.",
		},
	},
	{
		ID:       "algo-lru-cache",
		Category: "Interview Algorithms",
		Title:    "LRU Cache (Design Classic)",
		Description: `
<p>"Design an LRU cache" is the most-asked Go interview design problem.
The answer: a <b>doubly linked list</b> (O(1) move-to-front and
remove-tail) + a <b>map</b> (O(1) lookup).</p>

<p>Go ships <code>container/list</code>, which IS a doubly linked list. Use it.</p>

<ul>
  <li><b>Get(key):</b> map lookup; on hit, move that node to the front; return value.</li>
  <li><b>Put(key, value):</b> if exists, update + move to front. Else add to front; if over capacity, drop tail and remove from map.</li>
</ul>
`,
		Code: `package main

import (
	"container/list"
	"fmt"
)

type entry struct {
	key   int
	value int
}

type LRU struct {
	cap   int
	items map[int]*list.Element // key -> node holding *entry
	order *list.List            // front = most-recently-used
}

func NewLRU(capacity int) *LRU {
	return &LRU{
		cap:   capacity,
		items: make(map[int]*list.Element, capacity),
		order: list.New(),
	}
}

func (c *LRU) Get(key int) (int, bool) {
	el, ok := c.items[key]
	if !ok {
		return 0, false
	}
	c.order.MoveToFront(el)
	return el.Value.(*entry).value, true
}

func (c *LRU) Put(key, value int) {
	if el, ok := c.items[key]; ok {
		el.Value.(*entry).value = value
		c.order.MoveToFront(el)
		return
	}
	if c.order.Len() == c.cap {
		oldest := c.order.Back()
		if oldest != nil {
			c.order.Remove(oldest)
			delete(c.items, oldest.Value.(*entry).key)
		}
	}
	el := c.order.PushFront(&entry{key: key, value: value})
	c.items[key] = el
}

func main() {
	c := NewLRU(2)
	c.Put(1, 10)
	c.Put(2, 20)
	fmt.Println(c.Get(1)) // 10 true   (1 is now MRU)
	c.Put(3, 30)          // evicts key 2 (LRU)
	_, ok := c.Get(2)
	fmt.Println("key 2 present?", ok) // false
	fmt.Println(c.Get(3))             // 30 true
}
`,
		Notes: []string{
			"container/list stores any in Value — cast back to *entry inside operations.",
			"ALWAYS store the KEY inside the entry — needed when evicting to delete from the map.",
			"For concurrent access wrap with a sync.Mutex (or use hashicorp/golang-lru).",
			"Common follow-ups: TTL eviction (heap by expiry), 2-tier (memory + disk), distributed (Redis).",
		},
	},
}
