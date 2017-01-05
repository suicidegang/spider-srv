package snippets

var (
	PostgresSearchCleanup string = `
	DROP MATERIALIZED VIEW IF EXISTS spider_urls_index;
	`
	PostgresSearchIndexCreate string = `
	CREATE MATERIALIZED VIEW spider_urls_index AS
	SELECT id, url, title, description, 
	      setweight(to_tsvector('es', regexp_replace(coalesce(urls.url), '[^\w]+', ' ', 'gi')), 'A') || 
	    setweight(to_tsvector('es', coalesce(urls.title)), 'A') ||
	    setweight(to_tsvector('es', coalesce(urls.description)), 'B')  as document
	FROM spider_urls urls;
	`
	PostgresSearchIndexCreateFullText string = `
	CREATE INDEX idx_fts_search ON spider_urls_index USING gin(document);
	`
	PostgresSearchReloadIndex string = `
	REFRESH MATERIALIZED VIEW spider_urls_index;
	`
	PostgresSearch string = `
	SELECT id, title, url
	FROM spider_urls_index
	WHERE document @@ plainto_tsquery('spanish', ?1)
	ORDER BY ts_rank(document, plainto_tsquery('spanish', ?1)) DESC;
	`
)
