const prod = process.env.NODE_ENV === "production"

module.exports = {
    plugins: prod ? [
        require('autoprefixer'),
        require('cssnano'),
        require('postcss-preset-env'),
        require('pixrem'),
        require('postcss-pseudoelements'),
    ] : [],
}
