onload = () => {
	document.querySelectorAll(".copy").forEach((a) =>
		a.onclick = (event) =>
			event.preventDefault() |
			navigator.clipboard.writeText(a.href || a.dataset.href) |
			a.animate([{ background: "#C9F" }, {}], { duration: 1000 })
	);
};
