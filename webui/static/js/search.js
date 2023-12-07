const initNavigationSearch = () => {
  const openSearch = e => {
    e.preventDefault()
    let navigation = e.target.closest(".p-navigation")

    navigation.classList.add("has-search-open")
    navigation.querySelector(".p-search-box__input").focus()
    navigation.querySelectorAll(".js-search-button").forEach(searchButton => {
      searchButton.setAttribute("aria-pressed", true)
    })
  }

  const closeSearch = () => {
    document.querySelector(".p-navigation").classList.remove("has-search-open")
    document.querySelectorAll(".js-search-button").forEach(searchButton => {
      searchButton.removeAttribute("aria-pressed")
    })
  }

  document.querySelectorAll(".js-search-button").forEach(button => {
    button.addEventListener("click", e => {
      e.preventDefault()
      if (e.target.closest(".p-navigation").classList.contains("has-search-open")) {
        closeSearch()
      } else {
        closeSearch()
        openSearch(e)
      }
    })
  })

  const searchBox = document.querySelector(".p-search-box__input")
  searchBox.addEventListener("keyup", ({ key, target: { value } }) => {
    if (key === "Escape") {
      closeSearch()
      return
    }
    document.querySelectorAll(".drink").forEach(r => {
      const content = r.textContent.toLowerCase()
      r.style.display = content.includes(value.toLowerCase()) ? "" : "none"
    })
  })

  document.querySelector(".p-navigation__search-overlay").addEventListener("click", closeSearch)
}
