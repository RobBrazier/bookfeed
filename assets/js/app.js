import Alpine from 'alpinejs'

window.Alpine = Alpine

Alpine.store('format', 'atom')

/**
 * @param {String} baseUrl
 * @param {bool} useInput
 */
Alpine.data('container', (baseUrl, useInput) => ({
	base: baseUrl,
	useInput: useInput,
	output() {
		let url = `${window.location.origin}/${this.base}`
		if (this.useInput) {
			url += `/${this.input}`
		}
		url += `.${this.$store.format}`
		let query = {}
		if (this.filter !== "") {
			query.filter = this.filter
		}
		const params = new URLSearchParams(query)
		if (params.size > 0) {
			url += `?${params.toString()}`
		}
		return url
	},
	input: '',
	filter: '',
}))

Alpine.start()
