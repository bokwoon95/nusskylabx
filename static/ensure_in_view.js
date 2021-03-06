"use strict";

// https://stackoverflow.com/a/37285344
function ensureInView(container, element) {
  if (!container || !element) {
    return;
  }

  // Determine container top and bottom
  const cTop = container.scrollTop;
  const cBottom = cTop + container.clientHeight;

  // Determine element top and bottom
  const eTop = element.offsetTop;
  const eBottom = eTop + element.clientHeight;

  // Check if out of view
  if (eTop < cTop) {
    container.scrollTop -= cTop - eTop;
  } else if (eBottom > cBottom) {
    container.scrollTop += eBottom - cBottom;
  }
}

document.addEventListener("DOMContentLoaded", function () {
  const sidebar = document.querySelector(".sidebar");
  const currentItem = document.querySelector(".current-item");
  ensureInView(sidebar, currentItem);
});
