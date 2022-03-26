const prod = process.env.NODE_ENV === "production"

/** @type {import("snowpack").SnowpackUserConfig } */
export default {
  mount: {
    public: { url: '/', static: true },
    src: { url: '/dist' },
  },
  plugins: [
    '@snowpack/plugin-typescript',
    ['@snowpack/plugin-sass', { compilerOptions: { embedSourceMap: true } }],
    '@snowpack/plugin-postcss',
    ['@snowpack/plugin-run-script', {
      cmd: 'copyfiles -f node_modules/@fortawesome/fontawesome-free/webfonts/* build/external/webfonts',
    }],
  ],
  optimize: {
    minify: prod,
  },
  buildOptions: {
    metaUrlPath: 'external',
    sourcemap: !prod,
    watch: !prod,
  },
}
