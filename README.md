# Lambda GO triggada com Bucket S3 para processar no DynamoDB [![My Skills](https://skillicons.dev/icons?i=go,terraform,aws)](https://skillicons.dev)  

![Image](https://github.com/user-attachments/assets/94c4c568-fc48-4975-8745-4005fdde1f0d)

### Funcionalidades
<p>-Processamento de arquivos excel que chegam em um Bucket S3. O insert dos dados é feito em um DynamoDB.</p>
<p>-O processamento está sendo feito usando a estratégia de bulk insert (em lote) e com alguns workers concorrendo (goroutines e channel)</p>
<p>-Provisionamento AWS com <b>Terraform</b> (API Gateway, Lambda, RDS, Bucket S3)</p>
<p>-CI/CD com Github Actions</p>


### Benchmarks
- Arquivo excel 250k registros sendo processado em 31s.
