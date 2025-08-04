# Lambda GO triggada com Bucket S3 para processar no DynamoDB [![My Skills](https://skillicons.dev/icons?i=go,terraform,aws)](https://skillicons.dev)  

![Image](https://github.com/user-attachments/assets/317dcf47-ca22-4dc2-824f-880b2f770699)

### Funcionalidades
<p>-Processamento de arquivos excel que chegam em um Bucket S3. O insert dos dados é feito em um DynamoDB.</p>
<p>-O processamento está sendo feito usando a estratégia de bulk insert (em lote) e com alguns workers concorrendo (goroutines e channel)</p>
<p>-Provisionamento recursos AWS com <b>Terraform</b> (API Gateway, Lambda, DynamobDB, Bucket S3)</p>
<p>-CI/CD com Github Actions</p>

### Cenários
- Cenário 1: Para arquivos menores que 10MB, o cliente pode mandar direto pelo API Gateway para salvar no S3, pelo path /upload/{nomeArquivo}.
- Cenário 2: Para arquivos maiores que 10MB, é necessário chamar um outro endpoint /presigned-url?fileName={nomeArquivo}. Esse endpoint irá invocar uma lambda que irá retornar uma URL Pré-assinada do Bucket S3. Tendo a URL Pré-assinada, o usuário pode enviar o arquivo maior que 10MB por ela, já que com ela não tem limite de tamanho.


### Benchmarks
- Arquivo excel com 1 milhão de registros sendo processado em 1m40s.