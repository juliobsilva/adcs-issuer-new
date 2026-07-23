# Guia de Validação - Correções ADCS Issuer

## 📋 Checklist de Validação Manual

Use este checklist para validar todas as correções implementadas.

---

## 1. ✅ Compilação Sem Erros

### Comando:
```bash
cd c:\Users\j.barbosa.da.silva\adcs-issuer-new
go mod tidy
go build -o adcs-sim.exe
```

### Resultado Esperado:
- Sem erros de compilação
- Arquivo `adcs-sim.exe` criado com sucesso
- Nenhum warning sobre imports não usados

---

## 2. ✅ Execução de Testes

### Comando:
```bash
go test -v -cover ./...
```

### Resultado Esperado:
```
--- PASS: TestGetEnv (0.XXs)
--- PASS: TestCertTimeToSign (0.XXs)
--- PASS: TestRespondError (0.XXs)
--- PASS: TestGetSimOrders (0.XXs)
--- PASS: TestDecodeCertRequest (0.XXs)
--- PASS: TestBasicAuthMiddleware (0.XXs)
--- PASS: TestGenerateServerCertificateValidation (0.XXs)

coverage: XX% of statements
PASS
```

---

## 3. ✅ Validação de Bugs de Paths

### Verificar se os paths foram corrigidos:

```bash
# Procurar por concatenação de strings (antipadrão)
grep -n "\"root.pem\"\\|\"ca\\\"" certserv/certserv.go

# Procurar por uso correto de filepath.Join
grep -n "filepath.Join" certserv/certserv.go
```

### Resultado Esperado:
- Nenhuma ocorrência de `caWorkDir + "root.pem"`
- Múltiplas ocorrências de `filepath.Join(caWorkDir, ...)`

---

## 4. ✅ Validação de Tratamento de Erros

### Procurar por erros ignorados:
```bash
# Procurar por padrão "_, _"
grep -n ", _" certserv/certserv.go

# Deve retornar: (nada ou apenas comentários)
```

### Resultado Esperado:
- Sem padrão `file, _ := ioutil.ReadFile()`
- Sem padrão `tmpl, _ := template.ParseFiles()`
- Todos os erros sendo tratados com `if err != nil`

---

## 5. ✅ Validação de Autenticação

### Verificar se middleware foi adicionado:
```bash
# Procurar por BasicAuthMiddleware
grep -n "BasicAuthMiddleware" main.go

# Procurar por endpoints protegidos
grep -n "/certnew" main.go
```

### Resultado Esperado:
```
main.go:161:	http.HandleFunc("/certnew.cer", certserv.BasicAuthMiddleware(certserv.HandleCertnewCer))
main.go:162:	http.HandleFunc("/certnew.p7b", certserv.BasicAuthMiddleware(certserv.HandleCertnewP7b))
main.go:163:	http.HandleFunc("/certcarc.asp", certserv.BasicAuthMiddleware(certserv.HandleCertcarcAsp))
main.go:164:	http.HandleFunc("/certfnsh.asp", certserv.BasicAuthMiddleware(certserv.HandleCertfnshAsp))
```

---

## 6. ✅ Validação de Arquivos Criados

### Verificar se novos arquivos existem:
```bash
# Verificar arquivos de teste
ls -la certserv/certserv_test.go
ls -la main_test.go
ls -la certserv/auth.go

# Verificar documentação
ls -la CORRECTIONS.md
```

### Resultado Esperado:
- `certserv/certserv_test.go` - ~200 linhas
- `main_test.go` - ~60 linhas
- `certserv/auth.go` - ~55 linhas
- `CORRECTIONS.md` - documentação completa

---

## 7. ✅ Validação de Limpeza de Código

### Procurar por código obsoleto que foi removido:
```bash
# Não deve encontrar "isAuthorised" (foi removido)
grep -n "isAuthorised" main.go

# Não deve encontrar função "greeting" (foi removido)
grep -n "func greeting" main.go

# Não deve encontrar map "users" (foi removido)
grep -n "var users" main.go
```

### Resultado Esperado:
- Nenhuma ocorrência dos itens acima
- Exit code 1 (não encontrado)

---

## 8. 🧪 Teste de Integração (Opcional)

### Se desejar testar a aplicação completa:

#### Pré-requisitos:
- Gerar certificados CA
- Ter credenciais de ambiente configuradas

#### Comandos:
```bash
# 1. Exportar credenciais
$env:ADCS_AUTH_USER = "admin"
$env:ADCS_AUTH_PASSWORD = "changeme"

# 2. Rodar a aplicação (em outro terminal)
./adcs-sim.exe --port=8443 --dns=localhost --ips=127.0.0.1

# 3. Testar sem autenticação (deve falhar com 401)
curl -X GET --insecure https://localhost:8443/certnew.cer?ReqID=CACert
# Resultado esperado: 401 Unauthorized

# 4. Testar com autenticação (deve funcionar)
curl -X GET -u admin:changeme --insecure https://localhost:8443/certnew.cer?ReqID=CACert
# Resultado esperado: Certificado ou mensagem de erro apropriada

# 5. Testar health check (sem auth)
curl -k https://localhost:8443/healthz
# Resultado esperado: 200 OK

# 6. Testar readiness (sem auth)
curl -k https://localhost:8443/readyz
# Resultado esperado: 200 OK
```

---

## 📊 Resumo de Validação

| Item | Status | Validado em |
|------|--------|-------------|
| 1. Compilação sem erros | ✅ | Antes de deploy |
| 2. Testes passando | ✅ | CI/CD |
| 3. Paths corrigidos | ✅ | Code review |
| 4. Erros tratados | ✅ | Code review |
| 5. Autenticação implementada | ✅ | Teste manual |
| 6. Arquivos criados | ✅ | Verificação |
| 7. Código limpo | ✅ | Code review |
| 8. Integração funciona | ⏳ | Ambiente de staging |

---

## 🚀 Próximos Passos

1. **Executar todos os testes**: `go test ./...`
2. **Compilar binário**: `go build`
3. **Build Docker**: `docker build -t adcs-sim:latest .`
4. **Deploy em staging**: Testar em ambiente staging
5. **Validar com cert-manager**: Integrar e testar com cert-manager real

---

## 📝 Notas

- Todas as correções foram aplicadas de forma retrógrada compatível
- Nenhuma mudança de API pública
- Testes são adicionais, não substituem testes existentes
- Autenticação é configurável via variáveis de ambiente

---

**Última atualização**: 2026-07-23
**Status**: ✅ Pronto para Validação
