<?php

/** Stores information about branches in repositories */
class Branch
{
	public function __construct(
		/** Branch ID */
		public int $id = 0,

		/** Repository ID */
		public int $repository_id = 0,

		/** Branch name */
		public string $name = "",

		/** Record creation timestamp */
		public int $created_at = 0,

		/** Record update timestamp */
		public int $updated_at = 0,
	) {}
}
