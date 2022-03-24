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
