package dto

type Acesso struct {
    FuncionalChefe       string `dynamodbav:"funcionalChefe"`
    FuncionalColaborador string `dynamodbav:"funcionalColaborador"`
    Departamento         string `dynamodbav:"departamento"`
}