import coreWebVitals from 'eslint-config-next/core-web-vitals'
import typescript from 'eslint-config-next/typescript'

// eslint-config-next 16 ships flat configs, so they are imported directly.
// Going through @eslint/eslintrc's FlatCompat instead — the shape most guides
// still show — fails here with a circular-structure error while validating the
// legacy schema.
const config = [...coreWebVitals, ...typescript, { ignores: ['.next/**', 'out/**', 'node_modules/**'] }]

export default config
