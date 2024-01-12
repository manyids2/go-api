package main

const queryPackage = `
(package_clause 
	(package_identifier) @definition.package_clause)
`

const queryImports = `
((import_spec) @definition.import_spec)
`

const queryConstants = `
(source_file 
	(const_declaration
	(const_spec) @definition.const_spec))
`

const queryVariables = `
(source_file 
	(var_declaration
	(var_spec) @definition.var_spec))
`

const queryTypeInterfaces = `
((type_spec
	(type_identifier)
	(interface_type)) @definition.interface_type)
`

const queryTypeStructs = `
((type_spec
	(type_identifier)
	(struct_type)) @definition.struct_type)
`

const queryFunctions = `
((comment)* (function_declaration) @definition.function_declaration)
`

const queryMethods = `
((comment)* (method_declaration) @definition.method_declaration)
`
