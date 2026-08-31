package catalog

import (
	"ipw/internal/httpx"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Handler serves the public, read-only taxonomy endpoints.
type Handler struct{ store Store }

func NewHandler(store Store) *Handler { return &Handler{store: store} }

func (h *Handler) Register(app *fiber.App) {
	g := app.Group("/api/v1/catalog")
	g.Get("/categories", h.categories)
	g.Get("/skills", h.skills)
}

type categoryNode struct {
	ID       string         `json:"id"`
	Slug     string         `json:"slug"`
	Name     string         `json:"name"`
	Children []categoryNode `json:"children"`
}

type skillDTO struct {
	ID         string  `json:"id"`
	Slug       string  `json:"slug"`
	Name       string  `json:"name"`
	CategoryID *string `json:"categoryId"`
}

// categories returns the taxonomy as a nested tree.
func (h *Handler) categories(c *fiber.Ctx) error {
	cats, err := h.store.ListCategories(c.Context())
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.Map{"categories": buildTree(cats)})
}

func (h *Handler) skills(c *fiber.Ctx) error {
	list, err := h.store.ListSkills(c.Context(), c.Query("category"))
	if err != nil {
		return err
	}
	out := make([]skillDTO, len(list))
	for i, s := range list {
		out[i] = skillDTO{ID: s.ID.String(), Slug: s.Slug, Name: s.Name, CategoryID: uuidPtrString(s.CategoryID)}
	}
	return httpx.OK(c, fiber.Map{"skills": out})
}

func buildTree(cats []Category) []categoryNode {
	children := map[uuid.UUID][]categoryNode{}
	var roots []categoryNode
	// Two passes: nodes may reference a parent that appears later in the slice.
	byID := map[uuid.UUID]Category{}
	for _, c := range cats {
		byID[c.ID] = c
	}
	for _, c := range cats {
		node := categoryNode{ID: c.ID.String(), Slug: c.Slug, Name: c.Name, Children: []categoryNode{}}
		if c.ParentID == nil {
			roots = append(roots, node)
		} else {
			children[*c.ParentID] = append(children[*c.ParentID], node)
		}
	}
	for i := range roots {
		attachChildren(&roots[i], cats, children)
	}
	return roots
}

func attachChildren(node *categoryNode, cats []Category, children map[uuid.UUID][]categoryNode) {
	id, _ := uuid.Parse(node.ID)
	kids := children[id]
	if kids == nil {
		kids = []categoryNode{}
	}
	for i := range kids {
		attachChildren(&kids[i], cats, children)
	}
	node.Children = kids
}

func uuidPtrString(p *uuid.UUID) *string {
	if p == nil {
		return nil
	}
	s := p.String()
	return &s
}
