// @ts-ignore (typescript will not be able to find lazyload because it has no
// typescript declaration file, so we must ignore)
import LazyLoad from "vanilla-lazyload";

document.addEventListener("DOMContentLoaded", function () {
  const lazyLoadInstance = new LazyLoad({ elements_selector: ".lazy" });
  console.log(lazyLoadInstance);
});
