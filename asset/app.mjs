onload = () => {
	document.querySelectorAll("a.copy").forEach((a) =>
		a.onclick = (event) =>
			event.preventDefault() |
			navigator.clipboard.writeText(a.href) |
			a.animate([{ background: "#C9F" }, {}], { duration: 1000 })
	);
};
