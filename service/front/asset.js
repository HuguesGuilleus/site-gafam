onload = () => {
	const INNERTEXT = "innerText",
		qsaFor = (id, f) => document.querySelectorAll(id).forEach(f),
		DateTimeFormat = (opt) =>
			new Intl.DateTimeFormat(document.documentElement.lang, {
				dateStyle: "full",
				...opt,
			});

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

	qsaFor(
		"time",
		(time) =>
			time[INNERTEXT] = (/T/.test(time.dateTime)
				? DateTimeFormat({ timeStyle: "long" })
				: DateTimeFormat({}))
				.format(new Date(time.dateTime)),
	);
};
