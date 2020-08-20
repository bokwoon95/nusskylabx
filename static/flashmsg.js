"use strict";

const flashmsg = document.querySelector(".flashmsg");
if (flashmsg) {
  const removeFlashmsg = () => flashmsg.remove();
  flashmsg.addEventListener("click", removeFlashmsg);
  setTimeout(removeFlashmsg, 1500);
}
