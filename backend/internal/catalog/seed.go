package catalog

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type seedCategory struct {
	Slug   string
	Name   string
	Skills []struct{ Slug, Name string }
}

func sk(slug, name string) struct{ Slug, Name string } {
	return struct{ Slug, Name string }{slug, name}
}

// seedData is the curated starter taxonomy. Extend freely; Seed is idempotent
// (upsert by slug), so re-running only adds/updates.
var seedData = []seedCategory{
	{Slug: "development", Name: "Разработка", Skills: []struct{ Slug, Name string }{
		sk("go", "Go"), sk("python", "Python"), sk("typescript", "TypeScript"),
		sk("react", "React"), sk("vue", "Vue"), sk("nodejs", "Node.js"),
		sk("postgresql", "PostgreSQL"), sk("docker", "Docker"), sk("kubernetes", "Kubernetes"),
	}},
	{Slug: "design", Name: "Дизайн", Skills: []struct{ Slug, Name string }{
		sk("ui-ux", "UI/UX"), sk("figma", "Figma"), sk("branding", "Брендинг"),
		sk("illustration", "Иллюстрация"), sk("motion", "Моушн-дизайн"),
	}},
	{Slug: "writing", Name: "Тексты и перевод", Skills: []struct{ Slug, Name string }{
		sk("copywriting", "Копирайтинг"), sk("technical-writing", "Техническая документация"),
		sk("translation", "Перевод"), sk("editing", "Редактура"),
	}},
	{Slug: "marketing", Name: "Маркетинг", Skills: []struct{ Slug, Name string }{
		sk("seo", "SEO"), sk("smm", "SMM"), sk("ppc", "Контекстная реклама"),
		sk("email-marketing", "Email-маркетинг"), sk("analytics", "Веб-аналитика"),
	}},
	{Slug: "data", Name: "Данные и AI", Skills: []struct{ Slug, Name string }{
		sk("data-analysis", "Анализ данных"), sk("ml", "Machine Learning"),
		sk("llm", "LLM / промпт-инжиниринг"), sk("data-engineering", "Data Engineering"),
	}},
	{Slug: "video-audio", Name: "Видео и аудио", Skills: []struct{ Slug, Name string }{
		sk("video-editing", "Монтаж видео"), sk("sound-design", "Саунд-дизайн"),
		sk("voice-over", "Озвучка"),
	}},
}

// Seed writes the starter taxonomy. Idempotent. Runs quietly regardless of the
// configured GORM log level.
func Seed(ctx context.Context, store *PostgresStore) error {
	store = &PostgresStore{db: store.db.Session(&gorm.Session{Logger: logger.Discard})}
	for pos, c := range seedData {
		catID, err := store.UpsertCategory(ctx, Category{Slug: c.Slug, Name: c.Name, Position: pos})
		if err != nil {
			return err
		}
		cid := catID
		for _, s := range c.Skills {
			if err := store.UpsertSkill(ctx, Skill{CategoryID: &cid, Slug: s.Slug, Name: s.Name}); err != nil {
				return err
			}
		}
	}
	return nil
}
