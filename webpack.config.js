const path = require("path");
const Dotenv = require("dotenv-webpack");

module.exports = {
  mode: "development", // development by default, pass `--mode production` to override
  // pass `--watch` to make webpack watch the directory and auto-transpile on changes

  entry: {
    // Follow this format:
    // -------------------
    // "output_destination": "input_source"
    //
    // e.g.
    // If the entry is
    //    "/admins/form_edit": __dirname + "/app/admins/form_edit",
    // webpack will transpile
    //    __dirname + "/app/admins/form_edit.ts" into --> __dirname + "/static" + "/admins/form_edit.js"
    // where __dirname is the project root directory

    // app
    "/app/past_year_showcase": __dirname + "/app/past_year_showcase",

    // skylab
    "/skylab/form_edit": __dirname + "/app/skylab/form_edit",

    // admins
    "/admins/list_applications": __dirname + "/app/admins/list_applications",
    "/admins/list_cohorts": __dirname + "/app/admins/list_cohorts",
    "/admins/list_feedbacks": __dirname + "/app/admins/list_feedbacks",
    "/admins/list_forms": __dirname + "/app/admins/list_forms",
    "/admins/list_periods": __dirname + "/app/admins/list_periods",
  },
  output: {
    path: __dirname + "/static",
    filename: "[name].js",
  },
  devtool: "source-map",
  optimization: {
    splitChunks: {
      cacheGroups: {
        commons: {
          test: /[\\/]node_modules[\\/]/,
          name: "vendor",
          chunks: "initial",
        },
      },
    },
  },
  module: {
    rules: [
      {
        test: /\.(jsx?|tsx?)$/,
        exclude: /node_modules/,
        use: [
          {
            loader: "babel-loader",
            options: {
              presets: [
                [
                  "@babel/preset-env",
                  {
                    targets: {
                      browsers: [">0.25%", "not ie 11", "not op_mini all"],
                    },
                  },
                ],
                "@babel/preset-typescript",
              ],
            },
          },
        ],
      },
    ],
  },
  resolve: {
    extensions: ["*", ".js", ".jsx", ".ts", ".tsx"],
  },
  plugins: [new Dotenv({ path: path.resolve(__dirname, "./.env") })],
};
