<?php

/** Stores basic information about repositories */
class Repository
{
	public function __construct(
		/** Repository ID */
		public int $id = 0,

		/** Repository name */
		public string $name = "",

		/** Repository URL */
		public string $url = "",

		/** Record creation timestamp */
		public int $created_at = 0,

		/** Record update timestamp */
		public int $updated_at = 0,
	) {}
}
