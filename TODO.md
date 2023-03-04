# TODO:

## Top Priority
- `mongo.All(collectionName string)`
  - `mongo.All(collectionName string, i interface{})`? -> Receber uma interface para poder retornar no formato correto.
- `mongo.First(collectionName string, i interface{})`
- `mongo.UpdateOne(collectionName string, i interface{})`
- `mongo.DeleteOne(collectionName string, object interface{})`
- `mongo.Where(collectionName string, filter interface{})`?

---

## Upgrades
- `mongo.FindOrCreate("CollectionName", Object)`
- `mongo.UpdateMany("CollectionName", filterObject, newObject)`
- `mongo.DeleteMany("CollectionName", filterObject)`