/*
 * NetSchool
 *
 * The API for the NetSchool irTech project
 *
 * API version: 4.30.43656
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package swagger

type PrepareLoginForm struct {
	Countries []PrepareEmLoginFormCountries `json:"countries,omitempty"`
	States    []PrepareEmLoginFormCountries `json:"states,omitempty"`
	Provinces []PrepareEmLoginFormCountries `json:"provinces,omitempty"`
	Cities    []PrepareEmLoginFormCountries `json:"cities,omitempty"`
	Funcs     []PrepareEmLoginFormCountries `json:"funcs,omitempty"`
	Schools   []PrepareEmLoginFormCountries `json:"schools,omitempty"`
	Cid       int32                         `json:"cid,omitempty"`
	Sid       int32                         `json:"sid,omitempty"`
	Pid       int32                         `json:"pid,omitempty"`
	Cn        int32                         `json:"cn,omitempty"`
	Sft       int32                         `json:"sft,omitempty"`
	Scid      int32                         `json:"scid,omitempty"`
	Hlevels   *interface{}                  `json:"hlevels,omitempty"`
	Ems       *interface{}                  `json:"ems,omitempty"`
}