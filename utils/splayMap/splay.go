package splayMap

import (
	"golang.org/x/exp/constraints"
)

type Vertex[C constraints.Ordered, V any] struct {
	key   C
	value V
}

type node[C constraints.Ordered, V any] struct {
	left, right, parent *node[C, V]
	vertex              Vertex[C, V]
}

func zig[C constraints.Ordered, V any](v *node[C, V]) {
	p := v.parent
	v.parent = p.parent
	p.parent = v
	if p.left == v {
		if v.right != nil {
			v.right.parent = p
		}
		p.left, v.right = v.right, p
	} else {
		if v.left != nil {
			v.left.parent = p
		}
		p.right, v.left = v.left, p
	}
}

func zigZag[C constraints.Ordered, V any](v *node[C, V]) {
	zig(v)
	zig(v)
}

func zigZig[C constraints.Ordered, V any](v *node[C, V]) {
	p := v.parent
	zig(p)
	zig(v)
}

func splay[C constraints.Ordered, V any](v *node[C, V]) {
	for v.parent != nil {
		if v.parent.parent == nil {
			zig(v)
		} else if (v.parent.left == v) == (v.parent.parent.left == v.parent) {
			zigZig(v)
		} else {
			zigZag(v)
		}
	}
}

type SplayTree[C constraints.Ordered, V any] struct {
	root *node[C, V]
}

func (t *SplayTree[C, V]) find(v *Vertex[C, V]) {
	if t.root == nil {
		return
	}
	for {
		if t.root.vertex.key < v.key { // go right
			if t.root.right == nil {
				break
			}
			t.root = t.root.right
		} else if t.root.vertex.key > v.key { // go left
			if t.root.left == nil {
				break
			}
			t.root = t.root.left
		} else { // found equal
			break
		}
	}
	splay(t.root)
}

func (t *SplayTree[C, V]) split(v *Vertex[C, V]) (r *SplayTree[C, V]) {
	if t.root == nil {
		return &SplayTree[C, V]{}
	}
	for {
		if t.root.vertex.key < v.key { // go right
			if t.root.right == nil {
				break
			}
			t.root = t.root.right
		} else { // go left
			if t.root.left == nil {
				break
			}
			t.root = t.root.left
		}
	}
	splay(t.root)
	r = &SplayTree[C, V]{}
	if t.root.vertex.key < v.key {
		r.root = t.root.right
		t.root.right = nil
		if r.root != nil {
			r.root.parent = nil
		}
	} else {
		r.root = t.root.left
		t.root.left = nil
		if r.root != nil {
			r.root.parent = nil
		}
		t.root, r.root = r.root, t.root
	}
	return r
}

func (t *SplayTree[C, V]) merge(r *SplayTree[C, V]) {
	if t.root == nil || r.root == nil {
		if t.root == nil {
			t.root = r.root
		}
		return
	}

	for t.root.right != nil {
		t.root = t.root.right
	}
	for r.root.left != nil {
		r.root = r.root.left
	}

	splay(t.root)
	splay(r.root)

	if t.root != nil {
		t.root.right = r.root
	}
	if r.root != nil {
		r.root.parent = t.root
	}
}

func (t *SplayTree[C, V]) Insert(v *Vertex[C, V]) {
	tmp := t.split(v)
	if t.root == nil {
		t.root = &node[C, V]{
			parent: t.root,
			vertex: *v,
		}
	} else {
		t.root.right = &node[C, V]{
			parent: t.root,
			vertex: *v,
		}
	}
	t.merge(tmp)
}

func (t *SplayTree[C, V]) Erase(v *Vertex[C, V]) {
	tmp := t.split(v)
	for tmp.root.left != nil {
		tmp.root = tmp.root.left
	}
	splay(tmp.root)
	tmp.root = tmp.root.right
	tmp.root.parent = nil
	t.merge(tmp)
}

func (v *node[C, V]) walk(w []Vertex[C, V]) []Vertex[C, V] {
	if v == nil {
		return w
	}
	w = v.left.walk(w)
	w = append(w, v.vertex)
	w = v.right.walk(w)
	return w
}

func (t *SplayTree[C, V]) Print(w []Vertex[C, V]) []Vertex[C, V] {
	return t.root.walk(w)
}

func (t *SplayTree[C, V]) Clear() {
	t.root = nil
}

func (t *SplayTree[C, V]) AddNode(key C, value V) {
	t.Insert(&Vertex[C, V]{key, value})
}

func (t *SplayTree[C, V]) CheckNode(key C) bool {
	t.find(&Vertex[C, V]{key: key})
	if t.root.vertex.key == key {
		return true
	}
	return false
}

func (t *SplayTree[C, V]) ReturnNodeValue(key C) (V, bool) {
	t.find(&Vertex[C, V]{key: key})
	if t.root.vertex.key == key {
		return t.root.vertex.value, true
	}
	var _default V
	return _default, false
}
