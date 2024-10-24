onload = () => {
	const qsaFor = (id, f) => document.querySelectorAll(id).forEach(f);
	qsaFor(".copy", (a) =>
		a.onclick = (event) =>
			event.preventDefault() |
			navigator.clipboard.writeText(a.href || a.dataset.href) |
			a.animate([{ background: "#C9F" }, {}], { duration: 1000 }));

	qsaFor(
		"img.open",
		(img) =>
			img.onclick = (event) =>
				event.ctrlKey
					? event.preventDefault() | window.open(img.src)
					: 0,
	);
};
