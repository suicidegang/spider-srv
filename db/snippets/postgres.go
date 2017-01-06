package snippets

var (
	PostgresSearchCleanup string = `
	DROP MATERIALIZED VIEW IF EXISTS spider_urls_index;
	DROP TEXT SEARCH CONFIGURATION IF EXISTS es;
	`
	PostgresSearchIndexCreate string = `
	CREATE TEXT SEARCH CONFIGURATION es ( COPY = spanish );
	ALTER TEXT SEARCH CONFIGURATION es ALTER MAPPING
	FOR hword, hword_part, word WITH unaccent, spanish_stem;
	CREATE MATERIALIZED VIEW spider_urls_index AS
	SELECT id,
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
	SELECT u.*
	FROM spider_urls_index uix
	INNER JOIN spider_urls u ON u.id = uix.id
	WHERE document @@ plainto_tsquery('spanish', ?1)
	ORDER BY ts_rank(document, plainto_tsquery('spanish', ?1)) DESC;
	`
)
