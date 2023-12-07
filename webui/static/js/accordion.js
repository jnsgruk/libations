const initAccordions = () => {
  const toggleExpandedAccordion = (element, show) => {
    const target = document.getElementById(element.getAttribute("aria-controls"))
    if (target) {
      element.setAttribute("aria-expanded", show)
      target.setAttribute("aria-hidden", !show)
    }
  }

  // Setup all accordions on the page.
  const expanders = document.querySelectorAll(".drink-expand-button")
  expanders.forEach(accordionContainer => {
    // Set up an event listener on the container so that panels can be added
    // and removed and events do not need to be managed separately.
    accordionContainer.addEventListener("click", event => {
      let button = event.target.closest("button")

      let controls = button.getAttribute("aria-controls")
      let element = document.getElementById(controls)
      let isOpen = element.getAttribute("aria-hidden") === "true"
      element.setAttribute("aria-hidden", !isOpen)
      button.setAttribute("aria-expanded", !isOpen)
    })
  })
}
