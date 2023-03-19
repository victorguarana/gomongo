# TODO:

## Top Priority
- `mongo.Where(collectionName string, filter interface{})`?
- `mongo.Find(collectionName string, filter interface{})`?

---

## Upgrades
- `mongo.All(collectionName string, i interface{})`? -> Receber uma interface para poder retornar no formato correto.
- `mongo.FindOrCreate("CollectionName", Object)`
- `mongo.UpdateMany("CollectionName", filterObject, newObject)`
- `mongo.DeleteMany("CollectionName", filterObject)`