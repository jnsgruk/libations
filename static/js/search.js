const initNavigationSearch = () => {
  const openSearch = e => {
    e.preventDefault()
    let navigation = e.target.closest(".p-navigation")

    navigation.classList.add("has-search-open")
    navigation.querySelector(".p-search-box__input").focus()
    navigation.querySelectorAll(".js-search-button").forEach(searchButton => {
      searchButton.setAttribute("aria-pressed", true)
    })
    document.querySelector(".grid.p-strip").style.marginTop = "118px"
  }

  const closeSearch = () => {
    document.querySelector(".p-navigation").classList.remove("has-search-open")
    document.querySelectorAll(".js-search-button").forEach(searchButton => {
      searchButton.removeAttribute("aria-pressed")
    })
    document.querySelector(".grid.p-strip").style.marginTop = "48px"
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

  const resetButton = document.querySelector(".p-search-box__reset")
  resetButton.addEventListener("click", () => {
    document.querySelectorAll(".drink").forEach(r => {
      r.style.display = ""
    })
  })

  document.querySelector(".p-search-box").addEventListener("submit", e => {
    e.preventDefault()
    document.activeElement.blur()
  })

  document.querySelector(".p-navigation__search-overlay").addEventListener("click", closeSearch)
}
