# Neostr - A Graph-Native Eventstore for Nostr

`neostr` is an experimental nostr event store that caches relational data
between events in a queryable graph-native database, Neo4j.

## Testing

### Unit Tests

Run unit tests with:

```bash
go test
```

### Integration Tests

To run integration test with Neo4j, first make sure you have Docker Compose
installed, then start the test Neo4j instance (make sure you don't have another
neo4j server already running on your machine):

```bash
docker compose up -d
```

Then run the Neo4j integration tests:

```bash
go test -tags Neo4j
```
