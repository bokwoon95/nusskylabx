"use strict";

// https://plainjs.com/javascript/utilities/set-cookie-get-cookie-and-delete-cookie-5/
function getCookie(name) {
  var v = document.cookie.match("(^|;) ?" + name + "=([^;]*)(;|$)");
  return v ? v[2] : null;
}
function setCookie(name, value, days) {
  var d = new Date();
  d.setTime(d.getTime() + 24 * 60 * 60 * 1000 * days);
  document.cookie = name + "=" + value + ";path=/;expires=" + d.toGMTString();
}
const name = "toggle_mobile_sidebar";
const hide = "hide";
const show = "show";
document.addEventListener("DOMContentLoaded", function () {
  const toggle = document.querySelector("#toggle_mobile_sidebar");
  const sidebar = document.querySelector(".toggleable_sidebar");
  if (!toggle || !sidebar) {
    return;
  }
  const hideSidebar = function () {
    sidebar.classList.remove("dn", "db", "dn-l");
    sidebar.classList.add("dn", "db-l");
  };
  const showSidebar = function () {
    sidebar.classList.remove("dn", "db", "dn-l");
    sidebar.classList.add("db", "db-l");
  };
  const initialState = getCookie(name);
  if (initialState === hide) {
    hideSidebar();
  } else {
    showSidebar();
  }
  toggle.addEventListener("click", function () {
    const currentState = getCookie(name);
    if (!currentState || currentState === show) {
      setCookie(name, hide, 1);
      hideSidebar();
    } else if (currentState === hide) {
      setCookie(name, show, 1);
      showSidebar();
    }
  });
});
