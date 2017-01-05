CREATE MATERIALIZED VIEW spider_urls_index AS
SELECT id, url, title, description, 
      setweight(to_tsvector('es', regexp_replace(coalesce(urls.url), '[^\w]+', ' ', 'gi')), 'A') || 
    setweight(to_tsvector('es', coalesce(urls.title)), 'A') ||
    setweight(to_tsvector('es', coalesce(urls.description)), 'B')  as document
FROM spider_urls urls;

CREATE INDEX idx_fts_search ON spider_urls_index USING gin(document);

SELECT id, title, url
FROM spider_urls_index
WHERE document @@ plainto_tsquery('spanish', 'vas a volar iphone 6')
ORDER BY ts_rank(document, plainto_tsquery('spanish', 'vas a volar iphone 6')) DESC;


SELECT to_tsvector('simple', regexp_replace(coalesce(urls.url), '[^\w]+', ' ', 'gi')) as document
FROM spider_urls urls;

CREATE TEXT SEARCH CONFIGURATION es ( COPY = spanish );
ALTER TEXT SEARCH CONFIGURATION es ALTER MAPPING
FOR hword, hword_part, word WITH unaccent, spanish_stem;