// @ts-check

// 1. Import the Skeleton plugin
const { skeleton } = require('@skeletonlabs/tw-plugin');
const forms = require('@tailwindcss/forms');

/** @type {import('tailwindcss').Config} */
module.exports = {
	// 2. Opt for dark mode to be handled via the class method
	darkMode: 'class',
	content: [
		'./src/**/*.{html,js,svelte,ts}',
		// 3. Append the path to the Skeleton package
		require('path').join(require.resolve(
			'@skeletonlabs/skeleton'),
			'../**/*.{html,js,svelte,ts}'
		)
	],
	theme: {
		extend: {},
	},
	plugins: [
		// 4. Append the Skeleton plugin (after other plugins)
    forms,
    skeleton({
      themes: { preset: [{name: "wintry", enhancements: true}, {name: "crimson", enhancements: true}, {name: "seafoam", enhancements: true}, {name: "vintage", enhancements: true}, {name: "modern", enhancements: true}, {name: "rocket", enhancements: true}, {name: "skeleton", enhancements: true}, {name: "sahara", enhancements: true}, {name: "hamlindigo", enhancements: true}, {name: "gold-nouveau", enhancements: true}] }
  })
	]
}