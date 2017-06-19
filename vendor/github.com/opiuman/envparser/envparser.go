package envparser

import (
	"log"
	"os"
	"reflect"
	"sort"
	"strings"
)

type envVar struct {
	key   string
	value string
}

type envVars []envVar

func (es envVars) Len() int           { return len(es) }
func (es envVars) Swap(i, j int)      { es[i], es[j] = es[j], es[i] }
func (es envVars) Less(i, j int) bool { return es[i].key < es[j].key }

//An EnvParser reoks the env prefix, the env variables and config
type EnvParser struct {
	prefix    string
	unmarshal func(data []byte, v interface{}) error
	envs      envVars
}

//New returns a new EnvParser that reads the prefix, config
//and gathers/sorts the env variables
func New(prefix string, unmarshal func(data []byte, v interface{}) error) *EnvParser {
	log.SetPrefix(prefix + " ")
	ep := EnvParser{prefix: prefix, unmarshal: unmarshal}

	for _, env := range os.Environ() {
		e := strings.SplitN(env, "=", 2)
		ep.envs = append(ep.envs, envVar{e[0], e[1]})
	}
	sort.Sort(ep.envs)

	return &ep
}

//Parse loops through the env variables which overwrite the value of the config fields
//if there's match found in the "ENV to config field" convension like :
//PREFIX_AAA_BBB_CC => aaa.bbb.ccc
func (ep *EnvParser) Parse(config interface{}) error {
	confV := reflect.ValueOf(config)
	var err error
	for _, env := range ep.envs {
		envKey := env.key
		if strings.HasPrefix(envKey, strings.ToUpper(ep.prefix)+"_") {
			path := strings.Split(envKey, "_")
			err = ep.overwriteFields(confV, envKey, path[1:], env.value)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ep *EnvParser) overwriteFields(objValue reflect.Value, fullpath string, path []string, envValue string) error {
	//dereferencing pointer if any
	for objValue.Kind() == reflect.Ptr {
		if objValue.IsNil() {
			log.Fatalf("can not handle nil" + fullpath)
		}
		objValue = reflect.Indirect(objValue)
	}
	switch objValue.Kind() {
	case reflect.Struct:
		return ep.overwriteStruct(objValue, fullpath, path, envValue)
	case reflect.Map:
		return ep.overwriteMap(objValue, fullpath, path, envValue)
	case reflect.Interface:
		if objValue.NumMethod() == 0 {
			if !objValue.IsNil() {
				return ep.overwriteFields(objValue.Elem(), fullpath, path, envValue)
			}
			var m map[string]interface{}
			nestedValue := reflect.MakeMap(reflect.TypeOf(m))
			objValue.Set(nestedValue)
			return ep.overwriteMap(nestedValue, fullpath, path, envValue)
		}
	}
	return nil
}

func (ep *EnvParser) overwriteStruct(objValue reflect.Value, fullpath string, path []string, envValue string) error {
	// map of struct uppercased fields, panic on field name collision
	//for example: "Name" and "name" consider as collision
	upperFields := make(map[string]int)
	for i := 0; i < objValue.NumField(); i++ {
		structField := objValue.Type().Field(i)
		upper := strings.ToUpper(structField.Name)
		if _, ok := upperFields[upper]; ok {
			log.Fatalf("field name '%s' collision", structField.Name)
		}
		upperFields[upper] = i
	}

	fieldIndex, ok := upperFields[path[0]]
	if !ok {
		log.Printf("no field matches environment variable '%s'", fullpath)
		return nil
	}
	field := objValue.Field(fieldIndex)
	structField := objValue.Type().Field(fieldIndex)
	//one level env/field direct match
	if len(path) == 1 {
		fieldVal := reflect.New(structField.Type)
		err := ep.unmarshal([]byte(envValue), fieldVal.Interface())
		if err != nil {
			return err
		}
		field.Set(reflect.Indirect(fieldVal))
		return nil
	}
	//set empty object if fields is nil
	switch structField.Type.Kind() {
	case reflect.Map:
		if field.IsNil() {
			field.Set(reflect.MakeMap(structField.Type))
		}
	case reflect.Ptr:
		if field.IsNil() {
			field.Set(reflect.New(structField.Type))
		}
	}
	//overwrite the nested field
	err := ep.overwriteFields(field, fullpath, path[1:], envValue)
	if err != nil {
		return err
	}

	return nil
}

func (ep *EnvParser) overwriteMap(mapObjValue reflect.Value, fullpath string, path []string, envValue string) error {
	//map key has to be a string
	if mapObjValue.Type().Key().Kind() != reflect.String {
		log.Printf("env '%s' not support non-string map key ", fullpath)
		return nil
	}
	//loop overwrite the map values
	if len(path) > 1 {
		for _, k := range mapObjValue.MapKeys() {
			if strings.ToUpper(k.String()) == path[0] {
				mapValue := mapObjValue.MapIndex(k)
				if (mapValue.Kind() == reflect.Ptr || mapValue.Kind() == reflect.Interface ||
					mapValue.Kind() == reflect.Map) && mapValue.IsNil() {
					break
				}
				return ep.overwriteFields(mapValue, fullpath, path[1:], envValue)
			}
		}
	}

	var mapValue reflect.Value
	if mapObjValue.Type().Elem().Kind() == reflect.Map {
		mapValue = reflect.MakeMap(mapObjValue.Type().Elem())
	} else {
		mapValue = reflect.New(mapObjValue.Type().Elem())
	}
	if len(path) > 1 {
		err := ep.overwriteFields(mapValue, fullpath, path[1:], envValue)
		if err != nil {
			return err
		}
	} else {
		err := ep.unmarshal([]byte(envValue), mapValue.Interface())
		if err != nil {
			return err
		}
	}

	mapObjValue.SetMapIndex(reflect.ValueOf(strings.ToLower(path[0])), reflect.Indirect(mapValue))

	return nil
}
