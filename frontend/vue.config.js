module.exports = {
  devServer: {
    host: "0.0.0.0",
    port: 8081,
    allowedHosts: "all",
    headers: {
      "Access-Control-Allow-Origin": "*",
    },
    webSocketServer: false,
  },
};
