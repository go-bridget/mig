<?php

/** Stores asset information for each commit */
class Asset
{
	public function __construct(
		/** Asset ID */
		public int $id = 0,

		/** Commit ID */
		public int $commit_id = 0,

		/** Filename */
		public string $filename = "",

		/** File contents */
		public string $contents = "",

		/** Record creation timestamp */
		public int $created_at = 0,

		/** Record update timestamp */
		public int $updated_at = 0,
	) {}
}
