<?php

/** Stores asset information for each commit */
class Asset
{
	public function __construct(
		/** Asset ID */
		public $id ?int = 0;

		/** Commit ID */
		public $commit_id ?int = 0;

		/** Filename */
		public $filename ?string = "";

		/** File contents */
		public $contents ?string = "";

		/** Record creation timestamp */
		public $created_at ?int = 0;

		/** Record update timestamp */
		public $updated_at ?int = 0;
	) {}
}
