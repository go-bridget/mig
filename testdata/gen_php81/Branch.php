<?php

/** Stores information about branches in repositories */
class Branch
{
	public function __construct(
		/** Branch ID */
		public $id ?int = 0;

		/** Repository ID */
		public $repository_id ?int = 0;

		/** Branch name */
		public $name ?string = "";

		/** Record creation timestamp */
		public $created_at ?int = 0;

		/** Record update timestamp */
		public $updated_at ?int = 0;
	) {}
}
