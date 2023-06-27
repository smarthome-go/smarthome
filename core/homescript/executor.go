package homescript

type Executor struct {
}

// Resolves a Homescript module
// func (self *Executor) ResolveModule(id string) (string, string, bool, bool, map[string]homescript.Value, error) {
// 	moduleScopeAdditions, exists := builtinModules[id]
// 	if exists {
// 		return "", id, true, true, moduleScopeAdditions, nil
// 	}
//
// 	script, found, err := database.GetUserHomescriptById(id, self.Username)
// 	if !found || err != nil {
// 		return "", "", found, true, nil, err
// 	}
// 	return script.Data.Code, id, true, true, make(map[string]homescript.Value), nil
// }
