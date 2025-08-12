# Changelog

## Version v0.5.0 (2025-08-12)

### Features

- **redis:** add username, password, tls support for redis (#131) (b8021463)

### Chores and tidying

- **deps:** update module github.com/testcontainers/testcontainers-go to v0.38.0 (#124) (b27cd8b3)
- **deps:** update module github.com/vektra/mockery/v2 to v3 (#126) (a0dcf5c5)
- **deps:** update module flamingo.me/flamingo/v3 to v3.16.0 (#128) (3eab989a)
- migrate golangci-lint to v2 (#129) (52509340)
- **deps:** update module github.com/vektra/mockery/v2 to v2.53.0 (#116) (28d66ff0)

## Version v0.4.0 (2025-02-17)

### Features

- **309774:** replace Redis container images with Valkey (#119) (cb60630d)

### Chores and tidying

- **deps:** update module github.com/testcontainers/testcontainers-go to v0.35.0 (#115) (d9732717)
- **deps:** update module flamingo.me/flamingo/v3 to v3.13.0 (#114) (d140c972)
- **deps:** update module github.com/vektra/mockery/v2 to v2.50.4 (#113) (0d5ef208)
- **deps:** update module github.com/vektra/mockery/v2 to v2.50.0 (#110) (27d8f168)
- **deps:** update dependency go to v1.23.4 (#108) (0362c4d7)
- **deps:** update module golang.org/x/sync to v0.10.0 (#109) (a0c6d298)
- **deps:** update module flamingo.me/dingo to v0.3.0 (#112) (2a3be960)
- **deps:** update module github.com/testcontainers/testcontainers-go to v0.34.0 (#107) (25f25ff5)
- **deps:** update dependency go to v1.23.2 (#105) (8074b54e)
- **deps:** update module github.com/vektra/mockery/v2 to v2.46.3 (#106) (e9695b7e)
- **deps:** update module github.com/vektra/mockery/v2 to v2.46.1 (#104) (f38e4de1)
- **deps:** update module flamingo.me/flamingo/v3 to v3.11.0 (#102) (231d94fc)

## Version v0.3.5 (2024-09-18)

### Fixes

- **load:** pass deadline to the loader in context (#101) (ba58a0ac)

### Chores and tidying

- **deps:** update dependency go to v1.23.1 (#93) (ec072342)
- **deps:** update module flamingo.me/flamingo/v3 to v3.10.1 (#99) (ddd5c3ee)
- **deps:** update module github.com/vektra/mockery/v2 to v2.46.0 (#100) (aeba6ae5)
- bump minimum Go version to 1.22 (#98) (e2b612ac)
- **deps:** update module flamingo.me/flamingo/v3 to v3.10.0 (#97) (5912a72c)
- **deps:** update module github.com/testcontainers/testcontainers-go to v0.33.0 (#95) (887e0729)
- **deps:** update module github.com/vektra/mockery/v2 to v2.45.0 (#94) (15450b0c)

## Version v0.3.4 (2024-08-06)

### Documentation

- Update readme badges to correct module path (#74) (941d556f)

### Chores and tidying

- **deps:** update module github.com/vektra/mockery/v2 to v2.44.1 (#91) (0fff3aff)
- Replace github.com/pkg/errors with Go core error package (#89) (1b57f953)
- Always use newest golangci lint, fix linting issues (#90) (41cb9548)
- **deps:** update module flamingo.me/flamingo/v3 to v3.9.0 (#88) (09671670)
- **deps:** update dependency go to v1.22.5 (#87) (50316382)
- **deps:** update module github.com/testcontainers/testcontainers-go to v0.32.0 (#85) (30340a25)
- **deps:** update module github.com/vektra/mockery/v2 to v2.43.2 (#86) (e8cd3ea2)
- **deps:** update dependency go to v1.22.3 (#81) (3b176d6f)
- **deps:** update module github.com/vektra/mockery/v2 to v2.43.0 (#80) (50d0640c)
- **deps:** update golangci/golangci-lint-action action to v6 (#84) (3a1abd09)
- **deps:** update module flamingo.me/flamingo/v3 to v3.8.1 (#83) (fa5886a0)
- **deps:** update module golang.org/x/sync to v0.7.0 (#77) (8663c461)
- **deps:** update module github.com/vektra/mockery/v2 to v2.42.2 (#78) (a4c92ee0)
- **deps:** update module github.com/testcontainers/testcontainers-go to v0.30.0 (#79) (e183c74c)
- **deps:** update module github.com/vektra/mockery/v2 to v2.42.1 (#75) (5b580e78)

## Version v0.3.3 (2024-03-12)

### Chores and tidying

- **deps:** update module github.com/testcontainers/testcontainers-go to v0.29.1 (#68) (5493f251)
- switch to MIT licensing to streamline with other flamingo modules (#72) (d543ecc4)
- **deps:** update module github.com/stretchr/testify to v1.9.0 (#70) (b06ce82a)
- **deps:** update module github.com/gomodule/redigo to v1.9.2 (#69) (06ee66d8)
- **deps:** update module flamingo.me/flamingo/v3 to v3.8.0 (#65) (cbd10c13)
- **deps:** update module github.com/vektra/mockery/v2 to v2.42.0 (#66) (595968e1)
- **deps:** update golangci/golangci-lint-action action to v4 (#67) (1d42d9ce)

## Version v0.3.2 (2024-02-06)

### Ops and CI/CD

- add semanticore for hassle-free releases (#63) (8e0c09c1)
- bump golangci-lint to latest version (#61) (94f08868)

### Documentation

- Simplify readme, add section about custom backends (#40) (3df81dbf)

### Chores and tidying

- **deps:** update module golang.org/x/sync to v0.6.0 (#60) (68e879fa)
- **deps:** bump github.com/opencontainers/runc from 1.1.5 to 1.1.12 (#62) (de6245e3)
- **deps:** update module github.com/vektra/mockery/v2 to v2.40.1 (#57) (75b6753d)
- **deps:** update module github.com/testcontainers/testcontainers-go to v0.27.0 (#59) (2f5c6476)
- **deps:** update module github.com/vektra/mockery/v2 to v2.38.0 (#55) (45a87a1c)
- **deps:** update actions/setup-go action to v5 (#56) (11ecfa8b)
- **deps:** update module github.com/vektra/mockery/v2 to v2.36.1 (#54) (9ea52d7e)
- **deps:** update module golang.org/x/sync to v0.5.0 (#53) (a192c256)
- **deps:** update module github.com/vektra/mockery/v2 to v2.36.0 (#51) (74253b39)
- **deps:** update module github.com/testcontainers/testcontainers-go to v0.26.0 (#52) (e62ea8e2)
- **deps:** update module github.com/vektra/mockery/v2 to v2.35.2 (#49) (3788c930)
- **deps:** update module golang.org/x/sync to v0.4.0 (#50) (25224f06)

