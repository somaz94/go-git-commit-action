# Changelog

All notable changes to this project will be documented in this file.

## [v1.7.2](https://github.com/somaz94/go-git-commit-action/compare/v1.7.1...v1.7.2) (2026-03-25)

### Features

- delete linters ([c3cfff4](https://github.com/somaz94/go-git-commit-action/commit/c3cfff462b15d73532d80e446653f15998da5cf4))

### Bug Fixes

- add HTTP status validation, SHA format check, and remove raw response logging ([c653a64](https://github.com/somaz94/go-git-commit-action/commit/c653a6476971efcf07a2244583824e100de21e8e))
- Dockerfile ([63dd617](https://github.com/somaz94/go-git-commit-action/commit/63dd6176e71a229b66f0bdd22c8c5ca2238d051f))
- apache license -> mit license ([0078300](https://github.com/somaz94/go-git-commit-action/commit/0078300c1df9acb99907970dc43706717c44ae81))
- skip major version tag deletion on first release ([46f1e41](https://github.com/somaz94/go-git-commit-action/commit/46f1e41b7091b9b55290f4abbb21fb9345bbf6e5))

### Documentation

- add missing inputs to README table, optimize Docker build ([54782a8](https://github.com/somaz94/go-git-commit-action/commit/54782a85cfca42082e769b0836c1b94516898294))
- add no-push rule to CLAUDE.md ([8f324c5](https://github.com/somaz94/go-git-commit-action/commit/8f324c5b54d0f23559f348f29c5de4b4d930ef3e))
- update CLAUDE.md with commit guidelines and language ([7979814](https://github.com/somaz94/go-git-commit-action/commit/7979814215dcf3f33e71b3cda5d8cbcf6c73401e))

### Continuous Integration

- skip auto-generated changelog and contributors commits in release notes ([1d3416a](https://github.com/somaz94/go-git-commit-action/commit/1d3416a4cda2b53324568a2df5328feb41d05546))
- revert to body_path RELEASE.md in release workflow ([fba8198](https://github.com/somaz94/go-git-commit-action/commit/fba819825c7eb964f5dbca5f6ef635f2f3787a90))
- use generate_release_notes instead of RELEASE.md ([d59ba82](https://github.com/somaz94/go-git-commit-action/commit/d59ba82f88c1349e2f18c8394147eed1b1ff7ee6))
- migrate gitlab-mirror workflow to multi-git-mirror action ([550a4bc](https://github.com/somaz94/go-git-commit-action/commit/550a4bc1ced59131748d04c9879342f517dee7fd))
- use somaz94/contributors-action@v1 for contributors generation ([858b7fa](https://github.com/somaz94/go-git-commit-action/commit/858b7fac375dbb4afbdc04a862d1d29a637bf469))
- use major-tag-action for version tag updates ([764e283](https://github.com/somaz94/go-git-commit-action/commit/764e28375bcaf611f1dec992bef18de5aa2686c1))
- migrate changelog generator to go-changelog-action ([e25955f](https://github.com/somaz94/go-git-commit-action/commit/e25955f44b25f865f3283bb9292aced65666cdc8))
- add release config, contributors and dependabot auto-merge workflows ([32e08de](https://github.com/somaz94/go-git-commit-action/commit/32e08deddbc2b0790238a3ce47deaf446ce55c80))
- update Go version from 1.23 to 1.26 ([2eab530](https://github.com/somaz94/go-git-commit-action/commit/2eab53081accd7b53a387487ed14068ed5f5204b))
- unify changelog-generator with flexible tag pattern ([e63c86b](https://github.com/somaz94/go-git-commit-action/commit/e63c86b28ee8c26ac16dc1a3e60925faa14ce079))
- use conventional commit message in changelog-generator workflow ([c40a597](https://github.com/somaz94/go-git-commit-action/commit/c40a59781427eef8c5d4fad21eb20c2be7ee09ce))

### Chores

- upgrade Go version to 1.26 ([7a76830](https://github.com/somaz94/go-git-commit-action/commit/7a768303c94c7d5f9da156e61b216f8c49336042))
- change license from MIT to Apache 2.0 ([67cc84a](https://github.com/somaz94/go-git-commit-action/commit/67cc84addc6374d8004901b10c2d788b3777b51d))
- migrate devcontainer feature from devcontainers-contrib to devcontainers-extra ([f73b6e4](https://github.com/somaz94/go-git-commit-action/commit/f73b6e4c19c1d9782d4ffec437e62eea26553c82))

### Contributors

- somaz

<br/>

## [v1.7.1](https://github.com/somaz94/go-git-commit-action/compare/v1.7.0...v1.7.1) (2026-03-09)

### Bug Fixes

- resolve retry chdir path duplication and skip commit/push in PR dry run mode ([1a83cd5](https://github.com/somaz94/go-git-commit-action/commit/1a83cd5b5a07c6c86b0f34e81922b09c35551ead))

### Contributors

- somaz

<br/>

## [v1.7.0](https://github.com/somaz94/go-git-commit-action/compare/v1.6.3...v1.7.0) (2026-03-09)

### Features

- add action outputs, draft PR, reviewers/assignees, and bug fixes ([a63a2cc](https://github.com/somaz94/go-git-commit-action/commit/a63a2ccb7d05e1193ee9e4cb60ba50971b9cfeff))
- add action outputs, draft PR, and PR reviewers/assignees support ([bc884dc](https://github.com/somaz94/go-git-commit-action/commit/bc884dcc9cc49cfab6f8f68cb028833034096c25))
- add action outputs, draft PR, and PR reviewers/assignees support ([df02785](https://github.com/somaz94/go-git-commit-action/commit/df0278512a8e45ffcbd271979f7b62d46c926734))

### Code Refactoring

- remove emojis, fix API array response handling, and improve error safety ([ed0e977](https://github.com/somaz94/go-git-commit-action/commit/ed0e977562efde3de8bb36c989e1d9731871e4b1))
- remove emojis from output, fix panic and API response bug ([a95c69f](https://github.com/somaz94/go-git-commit-action/commit/a95c69f0836944e2cec4b11f6d2725153a002d65))
- extract GitHub API client and remove code duplication ([790f7e1](https://github.com/somaz94/go-git-commit-action/commit/790f7e1824d8b5394a860d704a949b2d6c4cb58f))

### Tests

- add tests for new packages and improve coverage ([8f21aa2](https://github.com/somaz94/go-git-commit-action/commit/8f21aa270d5161527489d6541fd6aae553513d36))
- add unit tests for git and pr packages, add unit-tests CI job ([9a9317c](https://github.com/somaz94/go-git-commit-action/commit/9a9317ce20ee860d3ff57262105b3a073183b77b))

### Builds

- **deps:** bump docker/setup-buildx-action from 3 to 4 ([b1a223f](https://github.com/somaz94/go-git-commit-action/commit/b1a223fa24d5bbd186d3ae4fd12cbf6a644c5b9c))
- **deps:** bump docker/build-push-action from 6 to 7 ([a2f66dd](https://github.com/somaz94/go-git-commit-action/commit/a2f66dd8c79fa5cfeea42596df915bb433d46662))
- **deps:** bump golang in the docker-minor group ([481167e](https://github.com/somaz94/go-git-commit-action/commit/481167ef6beb9a5c81ea19a57287d7431712ae3c))

### Chores

- add Makefile and unify project config with other repositories ([63181a8](https://github.com/somaz94/go-git-commit-action/commit/63181a8837358658ebc853bb4ff20af36f797e9a))
- unify workflow structure with other repositories ([8a70427](https://github.com/somaz94/go-git-commit-action/commit/8a70427f094d08a87f21abe70434b9f39917831e))
- stale-issues, issue-greeting ([7c56e38](https://github.com/somaz94/go-git-commit-action/commit/7c56e383d7bb356ff42b655eb03d024b535fdf46))

### Contributors

- somaz

<br/>

## [v1.6.3](https://github.com/somaz94/go-git-commit-action/compare/v1.6.2...v1.6.3) (2025-11-27)

### Code Refactoring

- errors, executor, git ([459f97c](https://github.com/somaz94/go-git-commit-action/commit/459f97cf7dede5efcc5b91842971553cf95ab577))

### Documentation

- DEVELOPMENT.md ([f7a4dcf](https://github.com/somaz94/go-git-commit-action/commit/f7a4dcfc5e19e398b7afefab300b8b8318db8e6f))

### Contributors

- somaz

<br/>

## [v1.6.2](https://github.com/somaz94/go-git-commit-action/compare/v1.6.1...v1.6.2) (2025-11-25)

### Code Refactoring

- pr ([10e5663](https://github.com/somaz94/go-git-commit-action/commit/10e56639cf0f06f04b0e9cddefb1c14e02f2cb11))

### Documentation

- all ([7caf54b](https://github.com/somaz94/go-git-commit-action/commit/7caf54bcd3260f4a95db6d7dd4439fd1a2e0de48))

### Chores

- ci.yml, use-action.yml ([29855fe](https://github.com/somaz94/go-git-commit-action/commit/29855fea504bde7acae57e10888a1e578603808a))
- ci.yml ([c37df17](https://github.com/somaz94/go-git-commit-action/commit/c37df17e80427d940d72207d7003c47160c12d13))
- use-action.yml ([60899ba](https://github.com/somaz94/go-git-commit-action/commit/60899ba3228050280f89bd513ee3cfd53c3baf01))
- use-action.yml ([00a2a8b](https://github.com/somaz94/go-git-commit-action/commit/00a2a8b8c3e8d833fa1334d136e091b37c470716))

### Contributors

- somaz

<br/>

## [v1.6.1](https://github.com/somaz94/go-git-commit-action/compare/v1.6.0...v1.6.1) (2025-11-25)

### Code Refactoring

- all code ([3adec2a](https://github.com/somaz94/go-git-commit-action/commit/3adec2ae4ce2a992fe3681fccdae5bc5cda20fa0))

### Documentation

- README.md, docs ([fe7e719](https://github.com/somaz94/go-git-commit-action/commit/fe7e7199e9e53bc0b9d15b3d471c2604f0aa3103))
- README.md ([cf0c7fe](https://github.com/somaz94/go-git-commit-action/commit/cf0c7fe3ec097e5fd683a41b7c3eb57004154171))
- README.md ([7c2a9a5](https://github.com/somaz94/go-git-commit-action/commit/7c2a9a54ee3ac5f91708340881371bee8ffe2ba4))

### Chores

- CHANGELOG.md ([9381d26](https://github.com/somaz94/go-git-commit-action/commit/9381d26ef933b9936eaca7e902a218cc1f3e5869))
- release.yml ([6cef410](https://github.com/somaz94/go-git-commit-action/commit/6cef4106ba3422409caf534799e38855f0bc4548))
- release.yml ([9efe287](https://github.com/somaz94/go-git-commit-action/commit/9efe287b2a3dc385417cf9deef08d5fcff03adac))
- release.yml ([68a0dd6](https://github.com/somaz94/go-git-commit-action/commit/68a0dd69de6fe2c7bf1c09495a835ecfb73e55ca))
- use-action.yml ([6062e4c](https://github.com/somaz94/go-git-commit-action/commit/6062e4c9016b83843336b8dfb06f11598c85e634))

### Contributors

- somaz

<br/>

## [v1.6.0](https://github.com/somaz94/go-git-commit-action/compare/v1.5.3...v1.6.0) (2025-11-24)

### Code Refactoring

- commit.go ([8f7d0cf](https://github.com/somaz94/go-git-commit-action/commit/8f7d0cf1fc59c5ebd07f522acc3942b3529ee6de))
- config.go, commit.go ([801fc17](https://github.com/somaz94/go-git-commit-action/commit/801fc17cdb40f2f000de3c75892eb185745a929d))

### Builds

- **deps:** bump actions/checkout from 5 to 6 ([62a20b0](https://github.com/somaz94/go-git-commit-action/commit/62a20b09de43833d957ae93130bc6798c24b0b9d))

### Chores

- release.yml ([b6dd6cd](https://github.com/somaz94/go-git-commit-action/commit/b6dd6cd27a5188a1c7dcec03fd604bf431911c57))
- workflows ([6034553](https://github.com/somaz94/go-git-commit-action/commit/6034553d784d106d9155659c9899bf05f4e8c993))
- ci.yml ([de712e4](https://github.com/somaz94/go-git-commit-action/commit/de712e427fb56a25a1b1bf1697d5d106d567d7a3))

### Contributors

- somaz

<br/>

## [v1.5.3](https://github.com/somaz94/go-git-commit-action/compare/v1.5.2...v1.5.3) (2025-10-28)

### Code Refactoring

- internal/git ([400b337](https://github.com/somaz94/go-git-commit-action/commit/400b337271e7c9901b113c05da71dd69a99f8c71))
- internal/git , add: internal/gitcmd ([6602359](https://github.com/somaz94/go-git-commit-action/commit/6602359c0746def331e8ee46367af7f37455dbfd))

### Contributors

- somaz

<br/>

## [v1.5.2](https://github.com/somaz94/go-git-commit-action/compare/v1.5.1...v1.5.2) (2025-10-28)

### Code Refactoring

- config.go ([7759689](https://github.com/somaz94/go-git-commit-action/commit/7759689b2aab88c500be69069e64146bf43597bb))

### Builds

- **deps:** bump actions/checkout from 4 to 5 ([872d523](https://github.com/somaz94/go-git-commit-action/commit/872d523208bddf11d4ade6069ce170f17d71d015))
- **deps:** bump golang in the docker-minor group ([d361ab1](https://github.com/somaz94/go-git-commit-action/commit/d361ab1e8cf1d8a3d9d1a6cb9d7bdd673b6c07f4))
- **deps:** bump super-linter/super-linter from 7 to 8 ([43ee54c](https://github.com/somaz94/go-git-commit-action/commit/43ee54c430991c6fcd417b12674e7017f393c414))

### Contributors

- somaz

<br/>

## [v1.5.1](https://github.com/somaz94/go-git-commit-action/compare/v1.5.0...v1.5.1) (2025-04-15)

### Bug Fixes

- ci.yml, use-action.yml ([4300f08](https://github.com/somaz94/go-git-commit-action/commit/4300f08b2e48f01023abe58748782aa320bd60b0))
- ci.yml, config.go, tag.go ([1020c02](https://github.com/somaz94/go-git-commit-action/commit/1020c02bd28bcb0b42977db4aa7cf8cc6ead5cb6))
- pr.go ([56d5696](https://github.com/somaz94/go-git-commit-action/commit/56d5696fc22bba4e012aae36e2984aaa705a6ee6))
- commit.go ([fa35523](https://github.com/somaz94/go-git-commit-action/commit/fa35523447bdd777b0ae294430dbfb7264a31efc))

### Contributors

- somaz

<br/>

## [v1.5.0](https://github.com/somaz94/go-git-commit-action/compare/v1.4.2...v1.5.0) (2025-04-14)

### Bug Fixes

- ci.yml, use-action.yml ([4833f38](https://github.com/somaz94/go-git-commit-action/commit/4833f3828bfbe229a7bc9280576703d9667a816c))
- pr.go, ci.yml ([c667bb5](https://github.com/somaz94/go-git-commit-action/commit/c667bb5479d81dfdfb3dfd90e0439897f76fb464))
- ci.yml, use-action.yml ([a2c42f9](https://github.com/somaz94/go-git-commit-action/commit/a2c42f93307651165e237a12f69682aa552c1a97))
- ci.yml ([7d9752f](https://github.com/somaz94/go-git-commit-action/commit/7d9752fa8555c08bdb6574659f35ac318d3a83c8))
- ci.yml ([a107960](https://github.com/somaz94/go-git-commit-action/commit/a1079602c9ea65186a06f60587c3b52af2cad5f4))
- ci.yml ([23fe098](https://github.com/somaz94/go-git-commit-action/commit/23fe098b200af5bfc42dc78e6b31676477366c9a))
- ci.yml ([4aabeeb](https://github.com/somaz94/go-git-commit-action/commit/4aabeebde43053f35efdecd4a101ab397c775714))
- ci.yml ([d2cc42c](https://github.com/somaz94/go-git-commit-action/commit/d2cc42cb992d0a319564ad576ba22431e4a98dae))
- ci.yml ([0a2a259](https://github.com/somaz94/go-git-commit-action/commit/0a2a259ddaee1a5dad18079b61f35c86f9e61007))
- ci.yml ([3464377](https://github.com/somaz94/go-git-commit-action/commit/34643770af819cb939884ea2250359199b3553af))
- ci.yml, pr.go ([005a919](https://github.com/somaz94/go-git-commit-action/commit/005a919be8c521a2597814d81a751241f0f7395d))
- ci.yml ([fc82501](https://github.com/somaz94/go-git-commit-action/commit/fc825019e9d9a97dc98d4667e4630049b2fafab7))
- ci.yml ([c198fe1](https://github.com/somaz94/go-git-commit-action/commit/c198fe18ff91a43a4c173772773a8986dbed7b05))
- ci.yml ([7c64687](https://github.com/somaz94/go-git-commit-action/commit/7c646876f6eb4acadad11e12822ce8544b582a08))
- ci.yml ([aa16b1a](https://github.com/somaz94/go-git-commit-action/commit/aa16b1a9737f54dffa95b7bd1c469a5f33f7005e))
- ci.yml, action.yml, config.go, pr.go ([7baa43d](https://github.com/somaz94/go-git-commit-action/commit/7baa43d063489b1d327db5165c37776995669385))
- ci.yml ([56bf1ec](https://github.com/somaz94/go-git-commit-action/commit/56bf1eca86dac9ea2dff3482881b0879d9892702))
- commit.go, ci.yml ([6d25aac](https://github.com/somaz94/go-git-commit-action/commit/6d25aac4843ed12336ae963c8ffeacbf59584253))

### Documentation

- README.md ([64a2227](https://github.com/somaz94/go-git-commit-action/commit/64a2227006bca269a1b6b695a790da90e27bd495))
- README.md ([fefd513](https://github.com/somaz94/go-git-commit-action/commit/fefd513d3def67d663ae158e067e09d6779b9c20))

### Tests

- commit JSON file with separate pattern ([a472270](https://github.com/somaz94/go-git-commit-action/commit/a472270f8aae6aadd678cd5f58835550ddcef007))
- commit multiple files with space-separated pattern ([4630269](https://github.com/somaz94/go-git-commit-action/commit/4630269ef9f7d364b583fb8ba61e334d3f7cf7d7))

### Add

- test/multi-pattern ([f2ed4f0](https://github.com/somaz94/go-git-commit-action/commit/f2ed4f0419de3305934cbeaf81c970105f848050))

### Contributors

- somaz

<br/>

## [v1.4.2](https://github.com/somaz94/go-git-commit-action/compare/v1.4.1...v1.4.2) (2025-03-04)

### Bug Fixes

- commit.go ([9de6d02](https://github.com/somaz94/go-git-commit-action/commit/9de6d02e994701dd8b6af884c2ba06abc80a2dc4))
- backup/commit.go.bak ([b10fcd1](https://github.com/somaz94/go-git-commit-action/commit/b10fcd10e016298533936669874e755274125dc8))
- pr.go ([fbbfcdb](https://github.com/somaz94/go-git-commit-action/commit/fbbfcdb90143a033fac99fe215e335dd93e57951))
- pr.go ([9e62700](https://github.com/somaz94/go-git-commit-action/commit/9e627000b879de84aa5c3d808acf63cfaa9de0ac))
- tag.go ([d890124](https://github.com/somaz94/go-git-commit-action/commit/d890124c8bf3726b6601fe516c23d4489346cc70))
- tag.go ([cb60810](https://github.com/somaz94/go-git-commit-action/commit/cb60810e2875e4be9a7c1ecc7df7cf9b37d828cf))

### Contributors

- somaz

<br/>

## [v1.4.1](https://github.com/somaz94/go-git-commit-action/compare/v1.4.0...v1.4.1) (2025-02-27)

### Bug Fixes

- commit.go ([2a025a3](https://github.com/somaz94/go-git-commit-action/commit/2a025a35c2d6ef879fe8af7d4f33096d84029178))
- changelog-generator.yml ([abd5fd6](https://github.com/somaz94/go-git-commit-action/commit/abd5fd63f507131ebaee3ed9532e729c346c55fe))
- ci.yml ([cadecff](https://github.com/somaz94/go-git-commit-action/commit/cadecfffebeba5595ccc31c63a74470cc0d82cb1))
- dependabot.yml ([810a851](https://github.com/somaz94/go-git-commit-action/commit/810a851e0002e53d9abf3f7d66f84b4b3c46ccf8))

### Add

- gitlab-mirror.yml ([46d5a60](https://github.com/somaz94/go-git-commit-action/commit/46d5a60a78a166762d5f3e24cc09a3ffe1da4e21))

### Contributors

- somaz

<br/>

## [v1.4.0](https://github.com/somaz94/go-git-commit-action/compare/v1.3.2...v1.4.0) (2025-02-17)

### Bug Fixes

- backup/* ([da2075e](https://github.com/somaz94/go-git-commit-action/commit/da2075ea5e487df66d68a02fd76542917c6d3348))
- changelog-generator.yml ([007e5de](https://github.com/somaz94/go-git-commit-action/commit/007e5de18a7ae70cb462d377345b44521871322a))

### Chores

- Code advancement ([762c6a0](https://github.com/somaz94/go-git-commit-action/commit/762c6a05fcb341d9b4a0f659845c13e356591be2))

### Contributors

- somaz

<br/>

## [v1.3.2](https://github.com/somaz94/go-git-commit-action/compare/v1.3.1...v1.3.2) (2025-02-17)

### Bug Fixes

- action.yml, commit.go, pr.go ([c26d740](https://github.com/somaz94/go-git-commit-action/commit/c26d7408b8a9b096d056eceeb37b61f41725034d))

### Contributors

- somaz

<br/>

## [v1.3.1](https://github.com/somaz94/go-git-commit-action/compare/v1.3.0...v1.3.1) (2025-02-17)

### Bug Fixes

- pr.go ([1a295a5](https://github.com/somaz94/go-git-commit-action/commit/1a295a52f8d2b41a18da9ef324392eb9206c5106))
- pr.go ([6527846](https://github.com/somaz94/go-git-commit-action/commit/6527846e8a0aa1dcb2c32e5b7d45ea4ad05a1264))
- ci.yml, use-action.yml ([6b2e6c4](https://github.com/somaz94/go-git-commit-action/commit/6b2e6c4d356c8bd4d5c3b7b51788160fa531534e))
- ci.yml ([dc66711](https://github.com/somaz94/go-git-commit-action/commit/dc66711615b01ce1ece43c66b9c1d598867ab67c))
- pr.go, use-action.yml ([6ede85a](https://github.com/somaz94/go-git-commit-action/commit/6ede85aa526f19c2b2b9f0ea5b10e0c175a031a0))
- use-action.yml, ci.yml ([15ec350](https://github.com/somaz94/go-git-commit-action/commit/15ec350d6d1e53b3e1c50052842e92e8c957d680))
- commit.go, pr.go ([8d474df](https://github.com/somaz94/go-git-commit-action/commit/8d474dfcf95fe76c393bb843c97fb6c2a20b5a6a))
- ci.yml ([054fbcb](https://github.com/somaz94/go-git-commit-action/commit/054fbcbcc56405c5ee6551044210197f3d9468f7))
- pr.go ([02274ea](https://github.com/somaz94/go-git-commit-action/commit/02274ea4bb23a320f9cf911e2baa634c5bbb7afe))
- pr.go ([02801f7](https://github.com/somaz94/go-git-commit-action/commit/02801f73b81f04e976ee1cf5d8150944c6b90fcf))
- pr.go ([663ec72](https://github.com/somaz94/go-git-commit-action/commit/663ec729af58f9847261576c40e5ac1cf54454bf))
- pr.go ([c28c333](https://github.com/somaz94/go-git-commit-action/commit/c28c33379120d7d384f7b9043a42917f09324eb7))
- ci.yml, use-action.yml ([fc48652](https://github.com/somaz94/go-git-commit-action/commit/fc48652fa6c14f290eded78b3fed20527ff0fb74))
- commit.go, pr.go ([e9340d5](https://github.com/somaz94/go-git-commit-action/commit/e9340d52f9379ea606014d71fe630408f8048a0c))
- ci.yml, use-action.yml, README.md ([56d9868](https://github.com/somaz94/go-git-commit-action/commit/56d9868f32cca735115d1e57c219cbbc14f21778))
- ci.yml, use-action.yml, README.md ([8d3f7b3](https://github.com/somaz94/go-git-commit-action/commit/8d3f7b3e68d60168f1400a4503e1d68fc95bd1fc))
- action.yml , README.md ([47f23a3](https://github.com/somaz94/go-git-commit-action/commit/47f23a3a9c814fd99b71e6d50ca4350a7eef8346))
- use-action.yml ([044cbe9](https://github.com/somaz94/go-git-commit-action/commit/044cbe91f9413265cc0678d137ebfcc229b4ff2d))
- use-action.yml ([3a413d0](https://github.com/somaz94/go-git-commit-action/commit/3a413d0400d7fe145c3b733763b96e8a661fad10))

### Builds

- **deps:** bump golang in the docker-minor group ([d92e6aa](https://github.com/somaz94/go-git-commit-action/commit/d92e6aa325dfb807aa9f052a03912e81941428cb))

### Contributors

- somaz

<br/>

## [v1.3.0](https://github.com/somaz94/go-git-commit-action/compare/v1.2.4...v1.3.0) (2025-02-17)

### Bug Fixes

- ci.yml ([4ffa8a7](https://github.com/somaz94/go-git-commit-action/commit/4ffa8a7c01b480a068d771ec5f2caf74ef232cab))
- config.go, pr.go ([2d5af31](https://github.com/somaz94/go-git-commit-action/commit/2d5af314138e2d2d74642e27b93878b67214757a))
- internal/* ([c7b52fc](https://github.com/somaz94/go-git-commit-action/commit/c7b52fc71ef00d54d09da4d7a36e6386eb5490dc))
- use-action.yml ([6a0fa7a](https://github.com/somaz94/go-git-commit-action/commit/6a0fa7aa6fb9920aea77857710d3343c6037155b))

### Documentation

- README.md ([3cdc0a4](https://github.com/somaz94/go-git-commit-action/commit/3cdc0a4661824b7511d5fce7fb775bee2492279e))

### Contributors

- somaz

<br/>

## [v1.2.4](https://github.com/somaz94/go-git-commit-action/compare/v1.2.3...v1.2.4) (2025-02-13)

### Bug Fixes

- Dockerfile ([7721385](https://github.com/somaz94/go-git-commit-action/commit/7721385c8aa746a17848b294312f7dee49c961ed))
- Dockerfile ([5e55acd](https://github.com/somaz94/go-git-commit-action/commit/5e55acd34eade543a98e5504280b102dfc05d45c))
- Dockerfile ([5f3e743](https://github.com/somaz94/go-git-commit-action/commit/5f3e7437d3b8b982c151cc382c9eb12a9ccb2df0))

### Contributors

- somaz

<br/>

## [v1.2.3](https://github.com/somaz94/go-git-commit-action/compare/v1.2.2...v1.2.3) (2025-02-12)

### Bug Fixes

- pr.go ([0a0d7d0](https://github.com/somaz94/go-git-commit-action/commit/0a0d7d068e67da68ef87b2d9b2703e52733edba9))

### Contributors

- somaz

<br/>

## [v1.2.2](https://github.com/somaz94/go-git-commit-action/compare/v1.2.1...v1.2.2) (2025-02-12)

### Bug Fixes

- action.yml ([c43b46f](https://github.com/somaz94/go-git-commit-action/commit/c43b46f296368da743bad0aab96c226bf841cb04))
- pr.go ([8e6d004](https://github.com/somaz94/go-git-commit-action/commit/8e6d004ba0a33d0622599ac803dad359057ee50c))
- pr.go & config.go ([2949e56](https://github.com/somaz94/go-git-commit-action/commit/2949e560cff0be4d2ff4041de59780679c644116))
- pr.go ([1c5b36e](https://github.com/somaz94/go-git-commit-action/commit/1c5b36e0b82ecc8f7657ff9a58951c384530e938))
- ci.yml & use-action.yml ([af4681e](https://github.com/somaz94/go-git-commit-action/commit/af4681e827314b8f772e1e1acf9afcfe56ba6aab))

### Contributors

- somaz

<br/>

## [v1.2.1](https://github.com/somaz94/go-git-commit-action/compare/v1.2.0...v1.2.1) (2025-02-12)

### Bug Fixes

- ci.yml ([afad4be](https://github.com/somaz94/go-git-commit-action/commit/afad4bec2f70d4c96b037ac5e4c8dd259baffa8e))

### Contributors

- somaz

<br/>

## [v1.2.0](https://github.com/somaz94/go-git-commit-action/compare/v1.1.4...v1.2.0) (2025-02-12)

### Bug Fixes

- use-action.yml ([66f43e4](https://github.com/somaz94/go-git-commit-action/commit/66f43e44ca4168f43c6648eb6c407201acef4856))
- use-action.yml ([c75b761](https://github.com/somaz94/go-git-commit-action/commit/c75b76144e165c7eb7b58ff338c0140aff9d8c32))
- use-action.yml ([2cd52be](https://github.com/somaz94/go-git-commit-action/commit/2cd52be01f97e59adae7990af2ff29c5630ce89a))
- pr.go ([0568c88](https://github.com/somaz94/go-git-commit-action/commit/0568c88254dcf6b51b62b49755d4f8ad71acb52c))
- ci.yml ([c1cbc30](https://github.com/somaz94/go-git-commit-action/commit/c1cbc30352f46b7af1678211e64340dff7acb954))
- ci.yml ([9e51902](https://github.com/somaz94/go-git-commit-action/commit/9e51902b0b9ad28e7408fe5f8b0958c76b80a7fa))
- pr.go ([ced6daa](https://github.com/somaz94/go-git-commit-action/commit/ced6daa3cb6eb417b68fda5b5547b01c6428941b))
- pr.go ([f3f34de](https://github.com/somaz94/go-git-commit-action/commit/f3f34de337d8465780dad7d7088557a8bdc245dd))
- pr.go, commit.go ([54a5979](https://github.com/somaz94/go-git-commit-action/commit/54a59793d676ed7d0806324f7cbcb63b0986ec95))
- pr.go, commit.go ([b06801b](https://github.com/somaz94/go-git-commit-action/commit/b06801bfc40a6c75bc1da7cd760ff0d470e3c391))
- pr.go ([5e42017](https://github.com/somaz94/go-git-commit-action/commit/5e420171472a752a178ffa0045f3ff90d513bcb0))
- commit.go ([9891b30](https://github.com/somaz94/go-git-commit-action/commit/9891b30717180267c00e7eaaa1b79eca9f4a1e4f))
- all file ([d838405](https://github.com/somaz94/go-git-commit-action/commit/d8384053eed677ae42f89197c65f5cf09608f789))
- rollback ([71f46e1](https://github.com/somaz94/go-git-commit-action/commit/71f46e14563830822a679f203f5167230d1852f2))
- main.go, commit.go, pr.go ([8f0066f](https://github.com/somaz94/go-git-commit-action/commit/8f0066fed4942bf052caddde782a878f7cc8e63e))
- main.go, commit.go, pr.go ([1e36693](https://github.com/somaz94/go-git-commit-action/commit/1e36693aeb30218d596fb735bb00bee44d2aba44))
- rollback ([291c5d9](https://github.com/somaz94/go-git-commit-action/commit/291c5d9d36fd2f8cf833efecca1318a66f5abc6a))
- commit.go ([bc025a0](https://github.com/somaz94/go-git-commit-action/commit/bc025a04e6bb344b80b0fc3fab08d40d733eed6e))
- main.go , commit,go , pr.go ([fd66902](https://github.com/somaz94/go-git-commit-action/commit/fd66902cc3f43d1af7e20ea01f7a75f8139b2fea))
- pr.go , commit.go ([1c3331f](https://github.com/somaz94/go-git-commit-action/commit/1c3331fb4a141d04aac16c46f110fc72feb4bb6a))
- config.go , pr.go ([6b6eda8](https://github.com/somaz94/go-git-commit-action/commit/6b6eda86f07749687a444d8df8e0094c586f754b))
- commit & pr.go ([43c1a80](https://github.com/somaz94/go-git-commit-action/commit/43c1a807b70b6618b3714a79eb8b6f397ee46e29))
- config.go & action.yml ([50f2215](https://github.com/somaz94/go-git-commit-action/commit/50f2215c1764fdbd3a79ca7ca5400d6550b70f7f))
- internal/git/* ([beeaea8](https://github.com/somaz94/go-git-commit-action/commit/beeaea83f09f2c7dadc4fade972b5678e28ebbcd))
- pr.go ([57247f0](https://github.com/somaz94/go-git-commit-action/commit/57247f0448cfe4537289eba78e15eeaa42055e39))
- ci.yml , pr.go ([6deda7d](https://github.com/somaz94/go-git-commit-action/commit/6deda7d30b443b350224913efd1c3177b4b0da13))
- ci.yml , pr.go ([d003184](https://github.com/somaz94/go-git-commit-action/commit/d003184170d42da63cd1318fb3889a46f91467e6))
- pr.go ([7d8816a](https://github.com/somaz94/go-git-commit-action/commit/7d8816a33d8c709ec6a0f026ab6f38bd0dc1531d))
- pr.go ([b845467](https://github.com/somaz94/go-git-commit-action/commit/b8454677f95dd8c7601516f605a2f2953b502de7))
- pr.go ([75240b1](https://github.com/somaz94/go-git-commit-action/commit/75240b152a4c781485176dac730a37cf58f3e7f8))
- pr.go ([4782c98](https://github.com/somaz94/go-git-commit-action/commit/4782c98f9346cdc16eaf157958f5b6b0bc154a71))
- pr.go ([c0440f3](https://github.com/somaz94/go-git-commit-action/commit/c0440f38d277d53438fcad27c8de2e6bf707e96f))
- pr.go ([e57d5ad](https://github.com/somaz94/go-git-commit-action/commit/e57d5ad8dee948fa3d3a0173b71d1472e2b84a79))
- action.yml ([1bdbb50](https://github.com/somaz94/go-git-commit-action/commit/1bdbb50e035d90971c72ee676d7a49978d854105))
- pr.go ([84a9bc9](https://github.com/somaz94/go-git-commit-action/commit/84a9bc964d0eac1cbd5b8ad9f52113e9b012b2a6))
- Dockerfile ([5af601d](https://github.com/somaz94/go-git-commit-action/commit/5af601dde85dcab88331da69512384b1a3ea7b87))
- ci.yml , pr.go ([44bc1a9](https://github.com/somaz94/go-git-commit-action/commit/44bc1a986c6e81bb67a433d25c2736028acf59c7))
- pr.go ([f90bc43](https://github.com/somaz94/go-git-commit-action/commit/f90bc4314d6b17bc8035a83d37588fdb52af0ac9))
- pr.go ([b2287b3](https://github.com/somaz94/go-git-commit-action/commit/b2287b30816be305cccade57db2a833adb89e7fb))
- pr.go ([45d635b](https://github.com/somaz94/go-git-commit-action/commit/45d635b3680a178d2a8418c3f15f73af96ddd69e))
- pr.go ([40960c8](https://github.com/somaz94/go-git-commit-action/commit/40960c85a053c8d0fba4f2e96800553cad7d2405))
- action.yml , config.go, pr.go, ci.yml ([edc7b8d](https://github.com/somaz94/go-git-commit-action/commit/edc7b8d0f9647705dc315c624ba2c7387254b2be))
- pr.go ([84606ca](https://github.com/somaz94/go-git-commit-action/commit/84606ca8cbe6909893eae957009d268e0c5383bd))
- pr.go ([f9f3a0e](https://github.com/somaz94/go-git-commit-action/commit/f9f3a0e64cf5f3b65cb0ae8dd1faaa761e7230df))
- pr.go ([2471703](https://github.com/somaz94/go-git-commit-action/commit/2471703ba3a693987ee9b92601c42b0150a35b47))
- pr.go ([379c824](https://github.com/somaz94/go-git-commit-action/commit/379c82443b050cbe450a63758f041102acddfa8c))
- pr.go ([139f47a](https://github.com/somaz94/go-git-commit-action/commit/139f47a4cfd05fdd27ca47a80c9e61205b071c28))
- commit.go ([17ca440](https://github.com/somaz94/go-git-commit-action/commit/17ca44008e8014dc846cf7b4821efc6c5239fe4b))
- ci.yml ([0433790](https://github.com/somaz94/go-git-commit-action/commit/04337905260a88da2bf55649bc56ccbc09d24ffc))
- commit.go ([b7fd038](https://github.com/somaz94/go-git-commit-action/commit/b7fd03817ca4db095c3adf6f67d836491a40a962))
- ci.yml & use-action.yml ([8d43662](https://github.com/somaz94/go-git-commit-action/commit/8d4366262dc771cd3a2ecbfd3997407604ae2950))
- ci.yml & pr.go ([5542165](https://github.com/somaz94/go-git-commit-action/commit/5542165dc8125150a0d14c5232094b9dad760ea6))
- commit.go, ci.yml, use-action.yml ([430d22f](https://github.com/somaz94/go-git-commit-action/commit/430d22f747d7687b97ffd11bd59c2fc3185ff762))
- commit.go ([b5e90f0](https://github.com/somaz94/go-git-commit-action/commit/b5e90f0f6ae4a2b21d12476bd660a60691609067))
- commit.go ([c3893ef](https://github.com/somaz94/go-git-commit-action/commit/c3893efc7330517f3ee9aeb4ba05c02aabc7145c))
- commit.go ([c4dbf86](https://github.com/somaz94/go-git-commit-action/commit/c4dbf86a9e2c5c07d729c55eaa26c456ed8b108d))
- commit.go ([b5953db](https://github.com/somaz94/go-git-commit-action/commit/b5953db67b425cf61259db0f51c8313436f8c0dd))
- commit.go ([1126305](https://github.com/somaz94/go-git-commit-action/commit/1126305a0e67f7c047f777636068fda21b2b56cb))
- commit.go ([66961b6](https://github.com/somaz94/go-git-commit-action/commit/66961b6d8061d85c6d7e661dfc1cb8f8271d6ee5))
- commit.go ([c6db749](https://github.com/somaz94/go-git-commit-action/commit/c6db74971f9e5bb29258819bfb00db78d091fd62))
- commit.go ([7f0e69f](https://github.com/somaz94/go-git-commit-action/commit/7f0e69fbdf1373b5f9e649e9cb3f00c4eb185fac))
- commit.go ([c3191a1](https://github.com/somaz94/go-git-commit-action/commit/c3191a1dcef73172eddece3d18a276d324cff5e2))
- ci.yml & commit.go ([6a82314](https://github.com/somaz94/go-git-commit-action/commit/6a82314aff13a3fdbb5bfe17196665d65d8adad1))
- pr.go ([5cf4937](https://github.com/somaz94/go-git-commit-action/commit/5cf4937767ed35dacfeae512340889761a1fcee3))
- ci.yml ([f9d3120](https://github.com/somaz94/go-git-commit-action/commit/f9d31202613d527091aeed41356b6366ef9a6ef2))
- pr.go ([3a3c93e](https://github.com/somaz94/go-git-commit-action/commit/3a3c93edb014c39816e70600bb70aa0f6766ac5b))
- pr.go ([a6220d3](https://github.com/somaz94/go-git-commit-action/commit/a6220d30a2b36c3320532f60e97d4948f6d49232))
- commit.go & pr.go ([54717b7](https://github.com/somaz94/go-git-commit-action/commit/54717b7b4dc014f7d0af6b91e221a1bf5ba408c8))
- action.yml & pr.go ([1f68070](https://github.com/somaz94/go-git-commit-action/commit/1f68070e8ab435476de34a09e0803fa4bb17c85c))
- pr.go ([25abd53](https://github.com/somaz94/go-git-commit-action/commit/25abd5363249c86ec6887c5f58e4e337f7239e55))
- commit.go ([12eada0](https://github.com/somaz94/go-git-commit-action/commit/12eada0aa82fba4ecffb67f91911e38c73d4c8b2))
- ci.yml & Dockerfile ([1827bba](https://github.com/somaz94/go-git-commit-action/commit/1827bbacb771328d08c759f33bd3d8999a150ab0))
- pr.go ([d395697](https://github.com/somaz94/go-git-commit-action/commit/d3956978ffbc9721e14fdaeb563f764ecd6946cd))

### Add

- function pr.go ([2163e68](https://github.com/somaz94/go-git-commit-action/commit/2163e680a9b8e4184ddde9dd7595febd6b0e47d5))

### Contributors

- somaz

<br/>

## [v1.1.4](https://github.com/somaz94/go-git-commit-action/compare/v1.1.3...v1.1.4) (2025-02-07)

### Bug Fixes

- use-action.yml ([9ecd062](https://github.com/somaz94/go-git-commit-action/commit/9ecd062589d68b40a7431b22c42490cdd702cf96))
- changelog-generator.yml ([5801e2d](https://github.com/somaz94/go-git-commit-action/commit/5801e2de20cb9ad98e6f61d6d96676669ded50b3))
- changelog-generator.yml ([1fb826b](https://github.com/somaz94/go-git-commit-action/commit/1fb826b09861436387fdfd5e8518f1f15e90d0cd))

### Contributors

- somaz

<br/>

## [v1.1.3](https://github.com/somaz94/go-git-commit-action/compare/v1.1.2...v1.1.3) (2025-02-07)

### Bug Fixes

- release.yml ([6c5ab02](https://github.com/somaz94/go-git-commit-action/commit/6c5ab0281955aa3d04c73ddb569fa742b20ea957))

### Contributors

- somaz

<br/>

## [v1.1.2](https://github.com/somaz94/go-git-commit-action/compare/v1.1.1...v1.1.2) (2025-02-07)

### Documentation

- README.md ([87be8b0](https://github.com/somaz94/go-git-commit-action/commit/87be8b0db052b62c4b4802d50572d89f7e03d2b4))

### Contributors

- somaz

<br/>

## [v1.1.1](https://github.com/somaz94/go-git-commit-action/compare/v1.1.0...v1.1.1) (2025-02-06)

### Bug Fixes

- file structure ([5cca3e8](https://github.com/somaz94/go-git-commit-action/commit/5cca3e846a1b1accddb708d00b811b800a14822c))

### Contributors

- somaz

<br/>

## [v1.1.0](https://github.com/somaz94/go-git-commit-action/compare/v1.0.2...v1.1.0) (2025-02-06)

### Bug Fixes

- use-action.yml ([9ec5598](https://github.com/somaz94/go-git-commit-action/commit/9ec559825dfcab491b47e4cf59f823de2fc3b7ea))
- release.yml ([20ac82e](https://github.com/somaz94/go-git-commit-action/commit/20ac82eecaa4b1fa34f271707e186c4d0396053a))
- use-action.yml ([60b23b5](https://github.com/somaz94/go-git-commit-action/commit/60b23b584500349db839141d6e8a9e717b91dc1c))
- release.yml ([b45b072](https://github.com/somaz94/go-git-commit-action/commit/b45b0726e4172257ce31284998a50b127d3ac948))
- ci.yml ([9739a9b](https://github.com/somaz94/go-git-commit-action/commit/9739a9b900fc8eb61bbdfb3f184fc01743a51d5c))
- ci.yml ([955aefb](https://github.com/somaz94/go-git-commit-action/commit/955aefb19de7336cce2a2109fc9da6f3bc8636d0))
- ci.yml ([a1f044f](https://github.com/somaz94/go-git-commit-action/commit/a1f044fc2eaab94a98bd5517ab489bbd5f766b2a))
- ci.yml ([efd1834](https://github.com/somaz94/go-git-commit-action/commit/efd18347bcd9c32fe7f360b2d965bec31b503dff))
- ci.yml ([81f91f7](https://github.com/somaz94/go-git-commit-action/commit/81f91f77ed1d63849325d2cdfa51467f04f0ac80))
- main.go ci.yml ([ab3c44a](https://github.com/somaz94/go-git-commit-action/commit/ab3c44a15104b8a31045f7e762769fb795f47d2f))
- main.go ([abccc53](https://github.com/somaz94/go-git-commit-action/commit/abccc53f783d1a6778556ce2af8055d2e8356ed9))
- action.yml main.go (add tag_reference) ([0b0daff](https://github.com/somaz94/go-git-commit-action/commit/0b0daffc6670e29f7e504794f3d6c279c0384399))

### Documentation

- README.md & fix: use-action.yml ([5354db5](https://github.com/somaz94/go-git-commit-action/commit/5354db5579776ebe84049bed0b9f2f1331a1f7fc))

### Contributors

- somaz

<br/>

## [v1.0.2](https://github.com/somaz94/go-git-commit-action/compare/v1.0.1...v1.0.2) (2025-02-06)

### Bug Fixes

- ci.yml & use-action.yml ([06fc27c](https://github.com/somaz94/go-git-commit-action/commit/06fc27cb884ca267a233492568d340c47abce22e))
- ci.yml ([ebd4783](https://github.com/somaz94/go-git-commit-action/commit/ebd4783d6738fb06da6d746aa3b30ff8d329cf7c))
- action.yml & main.go ([0c7b566](https://github.com/somaz94/go-git-commit-action/commit/0c7b56603b0a475b606e508c108461ee04c91b09))
- release.yml ([9e080a8](https://github.com/somaz94/go-git-commit-action/commit/9e080a863c7f0ed6c25a2f1ad5b7091091ba92b6))
- changelog-generator.yml & release.yml ([003d6f6](https://github.com/somaz94/go-git-commit-action/commit/003d6f67f67413ae91c0755284e5ceeb03a36133))
- changelog-generator.yml ([9a7ae9f](https://github.com/somaz94/go-git-commit-action/commit/9a7ae9f5055d2d21680c0c921c2c2b86b0800c62))
- changelog-generator.yml ([6f1f238](https://github.com/somaz94/go-git-commit-action/commit/6f1f23837c77baac5b4252171573ac9499667132))
- release.yml ([cffe065](https://github.com/somaz94/go-git-commit-action/commit/cffe06540303761c9ed81e9e7d23fa3f131fe3c6))
- release.yml ([72ddd22](https://github.com/somaz94/go-git-commit-action/commit/72ddd22e8a3132aff2ccfcf54b0e69d0546f60a8))

### Contributors

- somaz
- test1

<br/>

## [v1.0.1](https://github.com/somaz94/go-git-commit-action/compare/v1.0.0...v1.0.1) (2025-02-05)

### Bug Fixes

- changelog-generator.yml ([f8b8be3](https://github.com/somaz94/go-git-commit-action/commit/f8b8be3f4ee34686af64db0a1edc982fded7fa67))
- change-generator.yml ([a474a76](https://github.com/somaz94/go-git-commit-action/commit/a474a76cc2912c3cb882c9c7d2fc88ffb510af19))
- gitlab-ci.yml ([352818b](https://github.com/somaz94/go-git-commit-action/commit/352818b1c5ae8258a09e7fff1f38a4b251826008))
- release.yml ([577bfff](https://github.com/somaz94/go-git-commit-action/commit/577bfff9d384031fc008170430491034e5d4b0ea))
- changelog-generator.yml ([6667779](https://github.com/somaz94/go-git-commit-action/commit/66677797f128cececef3d2e04f707ee66439bfca))
- main.go ([fd97f8a](https://github.com/somaz94/go-git-commit-action/commit/fd97f8a972d023bf534149e4e73b1b7fe92204b9))
- use-action.yml ([5da3860](https://github.com/somaz94/go-git-commit-action/commit/5da3860c209b8dbf14dd76a12f8601b4fa1339bb))

### Documentation

- CODEOWNERS ([1cdf1e6](https://github.com/somaz94/go-git-commit-action/commit/1cdf1e61dafba31e9b6d20285712e125e3f85a32))
- README.md ([6579d4f](https://github.com/somaz94/go-git-commit-action/commit/6579d4f4a2bec6ef56da7ccab38f16deed6c0c44))
- README.md ([03cf418](https://github.com/somaz94/go-git-commit-action/commit/03cf418fe4aae61c65b1a6788862103a223bbb8c))

### Add

- release.yml & fix: changelog-generator.yml ([49e2222](https://github.com/somaz94/go-git-commit-action/commit/49e222287260d05be48c611299347712ea800bc4))

### Contributors

- somaz
- test1

<br/>

## [v1.0.0](https://github.com/somaz94/go-git-commit-action/releases/tag/v1.0.0) (2025-02-05)

### Bug Fixes

- move changelog-generator.yml ([72885e0](https://github.com/somaz94/go-git-commit-action/commit/72885e084a90325f63beb0b2b3af0bfbe3bb603a))
- ci.yml ([3d9bb54](https://github.com/somaz94/go-git-commit-action/commit/3d9bb54a9435f798337a2b169214f8cf1ff20aa1))
- action.yml & main.go ([4124e95](https://github.com/somaz94/go-git-commit-action/commit/4124e95a5da52dc49a17564beebd620447a4037a))
- ci.yml ([89a65ee](https://github.com/somaz94/go-git-commit-action/commit/89a65eefa115dd19d4790de70911235bb8d96f8e))
- main.go ([e5e6300](https://github.com/somaz94/go-git-commit-action/commit/e5e6300b8b97074d393b0ffec38e18c3106338e1))
- main.go ([768f19f](https://github.com/somaz94/go-git-commit-action/commit/768f19f68f42528c9c571e2e07a6b9bcd73bc51f))
- main.go ([b43c391](https://github.com/somaz94/go-git-commit-action/commit/b43c391f0373591232fac37797bec5688c559510))
- main.go ([2010935](https://github.com/somaz94/go-git-commit-action/commit/2010935dcce901fd5a67b9eff4601a446e96c242))
- action.yml & main.go ([4237960](https://github.com/somaz94/go-git-commit-action/commit/42379608ca651790a4cb95a817b162508a8f4002))
- main.go ([e4327af](https://github.com/somaz94/go-git-commit-action/commit/e4327af418da3c0889ea6473e2a53d03dc156d9e))
- ci.yml ([0f0926b](https://github.com/somaz94/go-git-commit-action/commit/0f0926b4f567b0d49dc8dd8da63f5884b1766251))
- ci.yml ([ba587c2](https://github.com/somaz94/go-git-commit-action/commit/ba587c2ca8498c7780627b8b1bb136051b144344))
- action.yml ([7ea0642](https://github.com/somaz94/go-git-commit-action/commit/7ea06424240836073515dd7334facbc24513a7d6))

### Add

- use-action.yml ([85c062e](https://github.com/somaz94/go-git-commit-action/commit/85c062e9cabd00684e888fafdc6c1f504fac2576))
- test/.gitkeep ([e46b235](https://github.com/somaz94/go-git-commit-action/commit/e46b235894cf7c2c5b3854e00332d55e833fab1a))

### Contributors

- somaz
- test1

<br/>

