Pod::Spec.new do |spec|

  spec.name         = "iden3"
  spec.version      = "0.0.1"
  spec.summary      = "SDK for using iden3 identity system."
  spec.description  = "SDK for using iden3 identity system."
  spec.homepage     = "https://iden3.io/"
  spec.license      = "GPL-3.0"
  spec.author             = { "iden3 Team" => "info@iden3.io" }
  spec.social_media_url   = "https://twitter.com/identhree"
  spec.platform     = :ios, "13.2"
  spec.swift_version = '5.0'
  spec.source       = { :path => '.' }
  spec.source_files  = "iden3"
  spec.exclude_files = "Classes/Exclude"
  spec.vendored_frameworks = 'iden3/Iden3mobile.framework'

end