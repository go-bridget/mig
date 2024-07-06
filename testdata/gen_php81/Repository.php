<?php

/** Stores basic information about repositories */
class Repository
{
	public function __construct(
		/** Repository ID */
		public $id ?int = 0;

		/** Repository name */
		public $name ?string = "";

		/** Repository URL */
		public $url ?string = "";

		/** Record creation timestamp */
		public $created_at ?int = 0;

		/** Record update timestamp */
		public $updated_at ?int = 0;
	) {}
}
