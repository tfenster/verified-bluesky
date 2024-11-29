$headers=@{}
$headers.Add("user-agent", "vscode-restclient")
$response = Invoke-WebRequest -Uri 'https://verified-bluesky-rpu7w3bv.fermyon.app/admin/' -Method GET -Headers $headers

$kvEntries = $response.Content | ConvertFrom-Json

# copy kvEntries to copiedKvEntries
Start-Sleep -seconds (60*60)
$copiedKvEntries = @()
$kvEntries | ForEach-Object {
    if ($_.key -ne "" -and $_.value -ne "") {
        $copiedKvEntries += $_
    }
}

$copiedKvEntries | ForEach-Object {
    if ($_.key -ne "" -and $_.value -ne "") {
        
        $key = $_.key
        $value = $_.value
        Write-Host "Key: $key, Value: $value"
        try {
            $response = Invoke-WebRequest -Uri 'https://verified-bluesky-rpu7w3bv.fermyon.app/admin/' -Method PUT -Headers $headers -Body ('{"Key": "'+ $key + '","Value": "' + $value + '"}')
            if ($response.StatusCode -eq 200) {
                Write-Host "Key: $key, Value: $value set successfully"
                $kvEntries = $kvEntries | ? {$_.key -ne $key}
            } else {
                Write-Host "Key: $key, Value: $value set failed"
            }
        } catch {
            Write-Host "Key: $key, Value: $value set failed"
        }
        Start-Sleep -seconds (5*60)
    }
}


